package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/resqlink-project/resqlink/internal/ai"
	"github.com/resqlink-project/resqlink/internal/domain"
	"github.com/resqlink-project/resqlink/internal/repository"
	"github.com/resqlink-project/resqlink/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	projectID := os.Getenv("GCP_PROJECT_ID")
	location := os.Getenv("GCP_LOCATION")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable is required")
	}
	if location == "" {
		location = "asia-south1"
	}

	subscriptionID := os.Getenv("PUBSUB_SUBSCRIPTION")
	if subscriptionID == "" {
		subscriptionID = "report-ingestion-sub"
	}

	// Firestore
	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore init: %v", err)
	}
	defer fsClient.Close()

	// Pub/Sub
	psClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("pubsub init: %v", err)
	}
	defer psClient.Close()

	// Gemini
	gemini, err := ai.NewGeminiClient(ctx, projectID, location)
	if err != nil {
		log.Fatalf("gemini init: %v", err)
	}
	defer gemini.Close()

	repo := repository.NewFirestoreRepo(fsClient)
	ingestionSvc := service.NewIngestionService(repo, gemini)

	sub := psClient.Subscription(subscriptionID)
	sub.ReceiveSettings.MaxOutstandingMessages = 10
	sub.ReceiveSettings.NumGoroutines = 4

	// Health check server for Cloud Run
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":    "ok",
				"service":   "resqlink-worker",
				"timestamp": time.Now().Unix(),
			})
		})
		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}
		log.Printf("worker health server on :%s", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Printf("health server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("received signal %v, shutting down", sig)
		cancel()
	}()

	log.Printf("RESQLINK worker listening on subscription: %s", subscriptionID)

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var event domain.IngestionEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("[worker] invalid message format: %v", err)
			msg.Ack()
			return
		}

		log.Printf("[worker] received report %s", event.ReportID)

		if err := ingestionSvc.ProcessReport(ctx, event); err != nil {
			log.Printf("[worker] processing failed for %s: %v", event.ReportID, err)
			msg.Nack()
			return
		}

		log.Printf("[worker] completed report %s", event.ReportID)
		msg.Ack()
	})

	if err != nil && ctx.Err() == nil {
		log.Fatalf("subscription receive error: %v", err)
	}

	log.Println("worker shutdown complete")
}
