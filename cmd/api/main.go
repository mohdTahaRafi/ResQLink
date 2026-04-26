package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	firebase "firebase.google.com/go/v4"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/samaj-project/samaj/internal/domain"
	"github.com/samaj-project/samaj/internal/middleware"
	"github.com/samaj-project/samaj/internal/repository"
	"github.com/samaj-project/samaj/internal/service"
)

func main() {
	ctx := context.Background()
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable is required")
	}

	// Firebase App
	fbApp, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		log.Fatalf("firebase init: %v", err)
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

	topic := psClient.Topic("report-ingestion")

	repo := repository.NewFirestoreRepo(fsClient)

	router := gin.Default()

	// CORS — allow Flutter web debug server to call the API
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check (unauthenticated)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	// Authenticated API group
	api := router.Group("/api/v1")
	api.Use(middleware.FirebaseAuth(fbApp))
	{
		api.POST("/reports", handleCreateReport(repo, topic))
		api.GET("/reports", handleListReports(repo))
		api.GET("/reports/:id", handleGetReport(repo))
		api.GET("/dashboard/:role", handleDashboard(repo))
		api.POST("/volunteers", handleCreateVolunteer(repo))
		api.GET("/match/:ward_id", handleMatchVolunteers(repo))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("SAMAJ API server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func handleCreateReport(repo *repository.FirestoreRepo, topic *pubsub.Topic) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("uid")

		var req struct {
			RawText   string  `json:"raw_text"`
			MediaURL  string  `json:"media_url"`
			MediaType string  `json:"media_type"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		report := &domain.Report{
			SubmitterUID: uid.(string),
			RawText:      req.RawText,
			MediaURL:     req.MediaURL,
			MediaType:    req.MediaType,
			Latitude:     req.Latitude,
			Longitude:    req.Longitude,
		}

		id, err := repo.CreateReport(c.Request.Context(), report)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create report"})
			log.Printf("create report error: %v", err)
			return
		}

		event := domain.IngestionEvent{
			ReportID:  id,
			MediaURL:  req.MediaURL,
			MediaType: req.MediaType,
			RawText:   req.RawText,
		}
		eventData, err := service.SerializeIngestionEvent(event)
		if err != nil {
			log.Printf("serialize event error: %v", err)
		} else {
			result := topic.Publish(c.Request.Context(), &pubsub.Message{Data: eventData})
			if _, err := result.Get(c.Request.Context()); err != nil {
				log.Printf("pubsub publish error: %v", err)
			}
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "status": "pending"})
	}
}

func handleListReports(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "50")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 50
		}

		reports, err := repo.GetAllReports(c.Request.Context(), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reports"})
			log.Printf("list reports error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"reports": reports, "count": len(reports)})
	}
}

func handleGetReport(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		report, err := repo.GetReport(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
			return
		}
		c.JSON(http.StatusOK, report)
	}
}

func handleDashboard(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.Param("role")

		reports, err := repo.GetAllReports(c.Request.Context(), 100)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard"})
			return
		}

		switch role {
		case "ngo_worker":
			c.JSON(http.StatusOK, gin.H{
				"mode":    "field",
				"reports": reports,
				"actions": []string{"submit_report", "view_queue", "sync_offline"},
			})
		case "lawyer":
			c.JSON(http.StatusOK, gin.H{
				"mode":    "legal",
				"reports": filterByCategory(reports, "legal"),
				"actions": []string{"search_fir", "search_deed", "case_history"},
			})
		case "clerk":
			pending := filterByStatus(reports, "pending")
			c.JSON(http.StatusOK, gin.H{
				"mode":          "digitization",
				"pending_queue": pending,
				"queue_size":    len(pending),
			})
		case "nagar_nigam":
			c.JSON(http.StatusOK, gin.H{
				"mode":        "command",
				"reports":     reports,
				"heatmap_data": buildHeatmapData(reports),
			})
		case "donor":
			resolved := filterByStatus(reports, "resolved")
			c.JSON(http.StatusOK, gin.H{
				"mode":             "transparency",
				"resolved_reports": resolved,
				"impact_count":     len(resolved),
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		}
	}
}

func handleCreateVolunteer(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		var vol domain.Volunteer
		if err := c.ShouldBindJSON(&vol); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		uid, _ := c.Get("uid")
		vol.UID = uid.(string)

		if vol.AssignedTasks > 0 {
			vol.CompletionRate = float64(vol.CompletedTasks) / float64(vol.AssignedTasks)
		}

		id, err := repo.CreateVolunteer(c.Request.Context(), &vol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register volunteer"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

func handleMatchVolunteers(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		wardID := c.Param("ward_id")

		ward, err := repo.GetWard(c.Request.Context(), wardID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ward not found"})
			return
		}

		volunteers, err := repo.GetVolunteersByWard(c.Request.Context(), wardID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch volunteers"})
			return
		}

		var skillsParam []string
		if skills := c.Query("skills"); skills != "" {
			json.Unmarshal([]byte(skills), &skillsParam)
		}
		if len(skillsParam) == 0 {
			skillsParam = []string{"general"}
		}

		results := service.MatchVolunteers(skillsParam, ward.CenterLat, ward.CenterLng, volunteers)

		topN := 10
		if len(results) < topN {
			topN = len(results)
		}

		c.JSON(http.StatusOK, gin.H{
			"ward_id":  wardID,
			"matches":  results[:topN],
			"total":    len(results),
			"returned": topN,
		})
	}
}

// --- helpers ---

func filterByCategory(reports []*domain.Report, cat string) []*domain.Report {
	var filtered []*domain.Report
	for _, r := range reports {
		if r.ProblemCategory == cat {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func filterByStatus(reports []*domain.Report, status string) []*domain.Report {
	var filtered []*domain.Report
	for _, r := range reports {
		if r.Status == status {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

type heatmapPoint struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Intensity float64 `json:"intensity"`
}

func buildHeatmapData(reports []*domain.Report) []heatmapPoint {
	points := make([]heatmapPoint, 0, len(reports))
	for _, r := range reports {
		if r.Latitude != 0 && r.Longitude != 0 {
			points = append(points, heatmapPoint{
				Lat:       r.Latitude,
				Lng:       r.Longitude,
				Intensity: r.UrgencyScore,
			})
		}
	}
	return points
}
