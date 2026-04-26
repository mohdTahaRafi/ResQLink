package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
)

func main() {
	ctx := context.Background()
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID not set")
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	topicID := "report-ingestion"
	subID := "report-ingestion-sub"

	// Create Topic
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatalf("Error checking topic: %v", err)
	}
	if !exists {
		topic, err = client.CreateTopic(ctx, topicID)
		if err != nil {
			log.Fatalf("Failed to create topic: %v", err)
		}
		fmt.Printf("Topic %v created.\n", topic.ID())
	} else {
		fmt.Printf("Topic %v already exists.\n", topicID)
	}

	// Create Subscription
	sub := client.Subscription(subID)
	exists, err = sub.Exists(ctx)
	if err != nil {
		log.Fatalf("Error checking subscription: %v", err)
	}
	if !exists {
		sub, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		if err != nil {
			log.Fatalf("Failed to create subscription: %v", err)
		}
		fmt.Printf("Subscription %v created.\n", sub.ID())
	} else {
		fmt.Printf("Subscription %v already exists.\n", subID)
	}
}
