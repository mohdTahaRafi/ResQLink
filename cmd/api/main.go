package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	firebase "firebase.google.com/go/v4"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/samaj-project/samaj/internal/ai"
	"github.com/samaj-project/samaj/internal/domain"
	"github.com/samaj-project/samaj/internal/middleware"
	"github.com/samaj-project/samaj/internal/repository"
	"github.com/samaj-project/samaj/internal/service"
)

// compressBase64Image decodes base64, subsamples, and re-encodes as JPEG
func compressBase64Image(b64 string) string {
	parts := strings.SplitN(b64, ",", 2)
	if len(parts) != 2 {
		return b64
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return b64
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return b64
	}
	b := img.Bounds()
	scale := 1
	if b.Dx() > 600 {
		scale = b.Dx() / 600
	}
	if b.Dy() > 600 && (b.Dy()/600) > scale {
		scale = b.Dy() / 600
	}
	if scale < 1 {
		scale = 1
	}
	var finalImg image.Image = img
	if scale > 1 {
		newB := image.Rect(0, 0, b.Dx()/scale, b.Dy()/scale)
		newImg := image.NewRGBA(newB)
		for y := 0; y < newB.Dy(); y++ {
			for x := 0; x < newB.Dx(); x++ {
				newImg.Set(x, y, img.At(x*scale+b.Min.X, y*scale+b.Min.Y))
			}
		}
		finalImg = newImg
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, finalImg, &jpeg.Options{Quality: 40})
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func main() {
	ctx := context.Background()
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable is required")
	}

	fbApp, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		log.Fatalf("firebase init: %v", err)
	}

	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore init: %v", err)
	}
	defer fsClient.Close()

	psClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("pubsub init: %v", err)
	}
	defer psClient.Close()

	repo := repository.NewFirestoreRepo(fsClient)
	topic := psClient.Topic("report-ingestion")
	authMW := middleware.FirebaseAuth(fbApp)

	// Initialize Gemini AI client
	location := os.Getenv("GCP_LOCATION")
	if location == "" {
		location = "asia-south1"
	}
	geminiClient, err := ai.NewGeminiClient(ctx, projectID, location)
	if err != nil {
		log.Printf("WARNING: Gemini AI client init failed (AI features disabled): %v", err)
	} else {
		defer geminiClient.Close()
		log.Println("✅ Gemini AI client initialized")
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	api := r.Group("/api/v1")
	api.Use(authMW)

	// ══════════════════════════════════════════════════
	// REPORTER — Submit Reports
	// ══════════════════════════════════════════════════
	api.POST("/reports", handleCreateReport(repo, topic, geminiClient))
	api.GET("/reports", handleListReports(repo))
	// IMPORTANT: static paths must be registered before parameterized ones
	api.GET("/reports/prioritized", handlePrioritizedReports(repo))
	api.GET("/reports/:id", handleGetReport(repo))
	api.GET("/reports/:id/match", handleMatchForReport(repo))
	api.PATCH("/reports/:id/status", handleUpdateReportStatus(repo))
	api.POST("/reports/:id/assign", handleAssignVolunteers(repo))

	// ══════════════════════════════════════════════════
	// VOLUNTEER — Task Management
	// ══════════════════════════════════════════════════
	api.GET("/volunteers/me/tasks", handleVolunteerTasks(repo))
	api.POST("/volunteers", handleCreateVolunteer(repo))

	// ══════════════════════════════════════════════════
	// SPECIALIST — Case Files & AI Search
	// ══════════════════════════════════════════════════
	api.GET("/cases/my", handleMyCases(repo))
	api.POST("/cases", handleCreateCase(repo))
	api.POST("/cases/:id/documents", handleAddDocument(repo))
	api.POST("/cases/:id/ask", handleAskCase(repo, geminiClient))
	api.POST("/cases/:id/search", handleSearchCase(repo, geminiClient))

	// ══════════════════════════════════════════════════
	// AI-POWERED FEATURES
	// ══════════════════════════════════════════════════
	aiGroup := api.Group("/ai")
	aiGroup.POST("/analyze-image", handleAnalyzeImage(geminiClient))
	aiGroup.POST("/verify-report", handleVerifyReport(geminiClient))
	aiGroup.POST("/detect-duplicates", handleDetectDuplicates(geminiClient, repo))
	aiGroup.POST("/action-plan", handleActionPlan(geminiClient, repo))
	aiGroup.POST("/sentiment", handleSentiment(geminiClient))
	aiGroup.POST("/translate", handleTranslate(geminiClient))
	aiGroup.GET("/progress-report", handleProgressReport(geminiClient, repo))
	aiGroup.POST("/recommend-skills", handleRecommendSkills(geminiClient))
	aiGroup.POST("/ocr", handleOCR(geminiClient))
	aiGroup.POST("/chat", handleAIChat(geminiClient))

	// ══════════════════════════════════════════════════
	// GEOCODING
	// ══════════════════════════════════════════════════
	api.POST("/geocode/reverse", handleReverseGeocode(geminiClient))

	// Legacy
	api.GET("/dashboard/:role", handleDashboard(repo))
	api.GET("/match/:ward_id", handleMatchVolunteers(repo))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("SAMAJ API server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// ══════════════════════════════════════════════════
// REPORTER Handlers
// ══════════════════════════════════════════════════

func handleCreateReport(repo *repository.FirestoreRepo, topic *pubsub.Topic, geminiClient *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("uid")

		var req struct {
			RawText            string  `json:"raw_text"`
			MediaURL           string  `json:"media_url"`
			MediaType          string  `json:"media_type"`
			Latitude           float64 `json:"latitude"`
			Longitude          float64 `json:"longitude"`
			IssueType          string  `json:"issue_type"`
			UserUrgency        string  `json:"user_urgency"`
			RequiredVolunteers int     `json:"required_volunteers"`
			Location           string  `json:"location"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if req.MediaURL != "" && strings.HasPrefix(req.MediaURL, "data:image") {
			req.MediaURL = compressBase64Image(req.MediaURL)
		}
		if len(req.MediaURL) > 1000000 {
			req.MediaURL = ""
			req.MediaType = "text"
			log.Println("WARNING: base64 image too large even after compression, dropped image.")
		}

		if req.IssueType == "" {
			req.IssueType = "civic_issue"
		}
		if req.UserUrgency == "" {
			req.UserUrgency = "normal"
		}
		if req.RequiredVolunteers < 1 {
			req.RequiredVolunteers = 1
		}
		if strings.TrimSpace(req.Location) == "" && (req.Latitude != 0 || req.Longitude != 0) {
			location, provider, err := reverseGeocode(c.Request.Context(), geminiClient, req.Latitude, req.Longitude)
			if err != nil {
				log.Printf("create report reverse geocode skipped: %v", err)
			} else {
				req.Location = location
				log.Printf("create report location resolved via %s: %s", provider, location)
			}
		}

		report := &domain.Report{
			SubmitterUID:       uid.(string),
			RawText:            req.RawText,
			MediaURL:           req.MediaURL,
			MediaType:          req.MediaType,
			Latitude:           req.Latitude,
			Longitude:          req.Longitude,
			IssueType:          req.IssueType,
			UserUrgency:        req.UserUrgency,
			RequiredVolunteers: req.RequiredVolunteers,
			Location:           req.Location,
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
			// Fire-and-forget: don't block the HTTP response
			result := topic.Publish(c.Request.Context(), &pubsub.Message{Data: eventData})
			go func() {
				if _, err := result.Get(context.Background()); err != nil {
					log.Printf("pubsub publish error (non-blocking): %v", err)
				}
			}()
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

// ══════════════════════════════════════════════════
// VOLUNTEER Handlers
// ══════════════════════════════════════════════════

func handleVolunteerTasks(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("uid")
		tasks, err := repo.GetReportsByAssignedVolunteer(c.Request.Context(), uid.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tasks"})
			log.Printf("volunteer tasks error: %v", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"tasks": tasks, "count": len(tasks)})
	}
}

func handleUpdateReportStatus(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Status string `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		validStatuses := map[string]bool{
			"pending": true, "accepted": true, "in_progress": true,
			"escalated": true, "resolved": true,
		}
		if !validStatuses[req.Status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}

		updates := []firestore.Update{{Path: "status", Value: req.Status}}
		if req.Status == "resolved" {
			updates = append(updates, firestore.Update{Path: "resolved_at", Value: time.Now()})
		}

		if err := repo.UpdateReport(c.Request.Context(), id, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": req.Status})
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

		// Force available=true so NGO can see this volunteer
		vol.Available = true

		// Set a default reliability score for new volunteers
		if vol.CompletionRate == 0 {
			vol.CompletionRate = 1.0 // new volunteers start with 100% reliability
		}

		// Check if volunteer already registered (prevent duplicates)
		existing, _ := repo.GetVolunteerByUID(c.Request.Context(), vol.UID)
		if existing != nil {
			log.Printf("Volunteer already registered: uid=%s id=%s", vol.UID, existing.ID)
			c.JSON(http.StatusOK, gin.H{"id": existing.ID, "existing": true})
			return
		}

		id, err := repo.CreateVolunteer(c.Request.Context(), &vol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register volunteer"})
			log.Printf("create volunteer error: %v", err)
			return
		}
		log.Printf("Volunteer registered: uid=%s id=%s name=%s skills=%v", vol.UID, id, vol.Name, vol.Skills)
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

// ══════════════════════════════════════════════════
// SPECIALIST Handlers
// ══════════════════════════════════════════════════

func handleMyCases(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("uid")
		cases, err := repo.GetCaseFilesBySpecialist(c.Request.Context(), uid.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch cases"})
			log.Printf("my cases error: %v", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"cases": cases, "count": len(cases)})
	}
}

func handleCreateCase(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Title         string   `json:"title"`
			ReportIDs     []string `json:"report_ids"`
			SpecialistUID string   `json:"specialist_uid"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		cf := &domain.CaseFile{
			Title:                 req.Title,
			ReportIDs:             req.ReportIDs,
			AssignedSpecialistUID: req.SpecialistUID,
		}
		id, err := repo.CreateCaseFile(c.Request.Context(), cf)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create case"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

func handleAddDocument(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		caseID := c.Param("id")
		var req struct {
			FileName string `json:"file_name"`
			Content  string `json:"content"`
			FileType string `json:"file_type"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		// Compress if image
		if strings.HasPrefix(req.Content, "data:image") {
			req.Content = compressBase64Image(req.Content)
		}
		if len(req.Content) > 900000 {
			req.Content = "" // strip if too large
		}

		doc := domain.CaseDocument{
			ID:         fmt.Sprintf("doc_%d", time.Now().UnixMilli()),
			FileName:   req.FileName,
			Content:    req.Content,
			FileType:   req.FileType,
			UploadedAt: time.Now(),
		}

		if err := repo.AddDocumentToCaseFile(c.Request.Context(), caseID, doc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add document"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"document_id": doc.ID})
	}
}

func handleAskCase(repo *repository.FirestoreRepo, geminiClient *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		caseID := c.Param("id")
		var req struct {
			Question string `json:"question"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		cf, err := repo.GetCaseFile(c.Request.Context(), caseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
			return
		}

		// Build context from all documents
		var contextBuilder strings.Builder
		for _, doc := range cf.Documents {
			contextBuilder.WriteString(fmt.Sprintf("--- Document: %s ---\n", doc.FileName))
			if doc.FileType == "text" || doc.FileType == "pdf" {
				contextBuilder.WriteString(doc.Content)
			} else {
				contextBuilder.WriteString("[Image document]")
			}
			contextBuilder.WriteString("\n\n")
		}

		// Also pull report texts
		for _, rid := range cf.ReportIDs {
			rpt, err := repo.GetReport(c.Request.Context(), rid)
			if err == nil {
				contextBuilder.WriteString(fmt.Sprintf("--- Report: %s ---\n%s\n\n", rid, rpt.RawText))
			}
		}

		docContext := contextBuilder.String()
		if docContext == "" {
			docContext = "No documents uploaded yet."
		}

		// Use Gemini AI for real AI-powered Q&A
		var answer string
		if geminiClient != nil {
			answer, err = geminiClient.AskCaseQuestion(c.Request.Context(), docContext, req.Question)
			if err != nil {
				log.Printf("Gemini case Q&A error: %v", err)
				answer = fmt.Sprintf(
					"AI analysis is temporarily unavailable. The case contains %d documents and %d linked reports.",
					len(cf.Documents), len(cf.ReportIDs))
			}
		} else {
			answer = fmt.Sprintf(
				"AI features are not enabled. The case contains %d documents and %d linked reports. "+
					"Please configure Gemini AI to enable intelligent Q&A.",
				len(cf.Documents), len(cf.ReportIDs))
		}

		c.JSON(http.StatusOK, gin.H{
			"answer":        answer,
			"case_id":       caseID,
			"doc_count":     len(cf.Documents),
			"context_chars": len(docContext),
		})
	}
}

func handleSearchCase(repo *repository.FirestoreRepo, geminiClient *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		caseID := c.Param("id")
		var req struct {
			Query string `json:"query"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		cf, err := repo.GetCaseFile(c.Request.Context(), caseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
			return
		}

		// Build document context
		var contextBuilder strings.Builder
		for _, doc := range cf.Documents {
			contextBuilder.WriteString(fmt.Sprintf("--- Document: %s ---\n%s\n\n", doc.FileName, doc.Content))
		}
		docContext := contextBuilder.String()

		// Use Gemini AI for semantic search if available
		if geminiClient != nil && docContext != "" {
			results, err := geminiClient.SemanticSearch(c.Request.Context(), docContext, req.Query)
			if err != nil {
				log.Printf("Gemini semantic search error: %v, falling back to keyword search", err)
			} else {
				c.JSON(http.StatusOK, gin.H{"results": results, "count": len(results)})
				return
			}
		}

		// Fallback: keyword search across documents
		var results []gin.H
		queryLower := strings.ToLower(req.Query)
		for _, doc := range cf.Documents {
			contentLower := strings.ToLower(doc.Content)
			if strings.Contains(contentLower, queryLower) {
				idx := strings.Index(contentLower, queryLower)
				start := idx - 100
				if start < 0 {
					start = 0
				}
				end := idx + len(queryLower) + 100
				if end > len(doc.Content) {
					end = len(doc.Content)
				}
				results = append(results, gin.H{
					"file_name": doc.FileName,
					"excerpt":   doc.Content[start:end],
					"score":     0.85,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{"results": results, "count": len(results)})
	}
}

func ensureAIAvailable(c *gin.Context, gemini *ai.GeminiClient) bool {
	if gemini != nil {
		return true
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error": "AI service unavailable. Configure GCP_PROJECT_ID, GCP_LOCATION, and Vertex AI credentials.",
	})
	return false
}

// ══════════════════════════════════════════════════
// GEOCODING Handler
// ══════════════════════════════════════════════════

func handleReverseGeocode(geminiClient *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body, need latitude and longitude"})
			return
		}

		if req.Latitude == 0 && req.Longitude == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "coordinates cannot both be zero"})
			return
		}
		if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "coordinates are outside valid latitude/longitude range"})
			return
		}

		location, provider, err := reverseGeocode(c.Request.Context(), geminiClient, req.Latitude, req.Longitude)
		if err != nil {
			log.Printf("Reverse geocode error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not determine location"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"location":  location,
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"provider":  provider,
		})
	}
}

func reverseGeocode(ctx context.Context, geminiClient *ai.GeminiClient, lat, lng float64) (string, string, error) {
	if apiKey := os.Getenv("GOOGLE_MAPS_API_KEY"); apiKey != "" {
		location, err := reverseGeocodeWithGoogle(ctx, apiKey, lat, lng)
		if err != nil {
			log.Printf("Google reverse geocode failed, falling back to Gemini: %v", err)
		} else if location != "" {
			return location, "google_maps", nil
		}
	}

	if geminiClient == nil {
		return "", "", fmt.Errorf("no reverse geocoding provider configured")
	}
	location, err := geminiClient.ReverseGeocode(ctx, lat, lng)
	return location, "gemini", err
}

func reverseGeocodeWithGoogle(ctx context.Context, apiKey string, lat, lng float64) (string, error) {
	values := url.Values{}
	values.Set("latlng", fmt.Sprintf("%.8f,%.8f", lat, lng))
	values.Set("key", apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://maps.googleapis.com/maps/api/geocode/json?"+values.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google geocode HTTP %d", resp.StatusCode)
	}

	var result struct {
		Status       string `json:"status"`
		ErrorMessage string `json:"error_message"`
		Results      []struct {
			FormattedAddress string `json:"formatted_address"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Status != "OK" {
		if result.ErrorMessage != "" {
			return "", fmt.Errorf("google geocode %s: %s", result.Status, result.ErrorMessage)
		}
		return "", fmt.Errorf("google geocode %s", result.Status)
	}
	if len(result.Results) == 0 || strings.TrimSpace(result.Results[0].FormattedAddress) == "" {
		return "", fmt.Errorf("google geocode returned no address")
	}
	return strings.TrimSpace(result.Results[0].FormattedAddress), nil
}

// ══════════════════════════════════════════════════
// NGO ADMIN Handlers
// ══════════════════════════════════════════════════

func handlePrioritizedReports(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		reports, err := repo.GetAllReports(c.Request.Context(), 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reports"})
			return
		}

		// Priority scoring: urgency weight × issue type weight × time decay
		type scoredReport struct {
			Report *domain.Report
			Score  float64
		}
		var scored []scoredReport
		now := time.Now()

		for _, r := range reports {
			urgencyWeight := 1.0
			switch r.UserUrgency {
			case "critical":
				urgencyWeight = 3.0
			case "urgent":
				urgencyWeight = 2.0
			}

			typeWeight := 1.0
			switch r.IssueType {
			case "medical_emergency":
				typeWeight = 3.0
			case "disaster_relief":
				typeWeight = 2.5
			case "legal_aid":
				typeWeight = 2.0
			}

			// Time bonus: older unresolved issues get higher priority
			hoursSince := now.Sub(r.CreatedAt).Hours()
			timeBonus := math.Log2(hoursSince + 1)

			score := urgencyWeight * typeWeight * (1 + timeBonus)
			if r.Status == "resolved" {
				score = 0 // resolved issues go to bottom
			}

			scored = append(scored, scoredReport{Report: r, Score: score})
		}

		// Sort by score descending
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].Score > scored[j].Score
		})

		result := make([]*domain.Report, len(scored))
		for i, s := range scored {
			result[i] = s.Report
		}

		c.JSON(http.StatusOK, gin.H{"reports": result, "count": len(result)})
	}
}

func handleMatchForReport(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		reportID := c.Param("id")
		report, err := repo.GetReport(c.Request.Context(), reportID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
			return
		}

		volunteers, err := repo.GetAllVolunteers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch volunteers"})
			return
		}

		// Map issue type to required skills
		skillsNeeded := []string{"general"}
		switch report.IssueType {
		case "medical_emergency":
			skillsNeeded = []string{"medical", "doctor", "paramedic", "general"}
		case "legal_aid":
			skillsNeeded = []string{"legal", "lawyer", "general"}
		case "disaster_relief":
			skillsNeeded = []string{"disaster", "rescue", "general"}
		}

		results := service.MatchVolunteers(skillsNeeded, report.Latitude, report.Longitude, volunteers)

		topN := 20
		if len(results) < topN {
			topN = len(results)
		}

		c.JSON(http.StatusOK, gin.H{
			"matches":  results[:topN],
			"total":    len(results),
			"returned": topN,
		})
	}
}

func handleAssignVolunteers(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		reportID := c.Param("id")
		var req struct {
			VolunteerIDs []string `json:"volunteer_ids"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		updates := []firestore.Update{
			{Path: "assigned_volunteer_ids", Value: req.VolunteerIDs},
			{Path: "status", Value: "accepted"},
		}
		if err := repo.UpdateReport(c.Request.Context(), reportID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"assigned": len(req.VolunteerIDs)})
	}
}

// ══════════════════════════════════════════════════
// LEGACY Handlers
// ══════════════════════════════════════════════════

func handleDashboard(repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.Param("role")
		reports, err := repo.GetAllReports(c.Request.Context(), 100)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"mode": role, "reports": reports, "count": len(reports)})
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
		c.JSON(http.StatusOK, gin.H{"ward_id": wardID, "matches": results[:topN], "total": len(results), "returned": topN})
	}
}

// ══════════════════════════════════════════════════
// AI FEATURE HANDLERS
// ══════════════════════════════════════════════════

// 1. Multimodal Image Analysis
func handleAnalyzeImage(gemini *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureAIAvailable(c, gemini) {
			return
		}
		var req struct {
			Image string `json:"image"` // base64 image
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Image == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image (base64) is required"})
			return
		}
		result, err := gemini.AnalyzeImage(c.Request.Context(), req.Image)
		if err != nil {
			log.Printf("AI image analysis error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "image analysis failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"analysis": result})
	}
}

// 2. AI Report Verification
func handleVerifyReport(gemini *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureAIAvailable(c, gemini) {
			return
		}
		var req struct {
			Image string `json:"image"` // base64
			Text  string `json:"text"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Image == "" || req.Text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image and text are required"})
			return
		}
		result, err := gemini.VerifyReport(c.Request.Context(), req.Image, req.Text)
		if err != nil {
			log.Printf("AI verify report error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "verification failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"verification": result})
	}
}

// 3. Duplicate Report Detection
func handleDetectDuplicates(gemini *ai.GeminiClient, repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureAIAvailable(c, gemini) {
			return
		}
		var req struct {
			Text string `json:"text"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
			return
		}
		// Fetch recent reports for comparison
		reports, err := repo.GetAllReports(c.Request.Context(), 20)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reports"})
			return
		}
		var existingSummaries []string
		for _, r := range reports {
			summary := r.RawText
			if r.Summary != "" {
				summary = r.Summary
			}
			existingSummaries = append(existingSummaries, summary)
		}
		result, err := gemini.DetectDuplicates(c.Request.Context(), req.Text, existingSummaries)
		if err != nil {
			log.Printf("AI duplicate detection error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "duplicate detection failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"duplicate_check": result})
	}
}

// 4. AI-Generated Action Plans
func handleActionPlan(gemini *ai.GeminiClient, repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureAIAvailable(c, gemini) {
			return
		}
		var req struct {
			ReportID string `json:"report_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.ReportID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "report_id is required"})
			return
		}
		report, err := repo.GetReport(c.Request.Context(), req.ReportID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
			return
		}
		result, err := gemini.GenerateActionPlan(c.Request.Context(), report)
		if err != nil {
			log.Printf("AI action plan error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "action plan generation failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"action_plan": result})
	}
}

// 5. Sentiment & Emotion Analysis
func handleSentiment(gemini *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureAIAvailable(c, gemini) {
			return
		}
		var req struct {
			Text string `json:"text"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
			return
		}
		result, err := gemini.AnalyzeSentiment(c.Request.Context(), req.Text)
		if err != nil {
			log.Printf("AI sentiment error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "sentiment analysis failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"sentiment": result})
	}
}

// 6. Real-Time Translation
func handleTranslate(gemini *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureAIAvailable(c, gemini) {
			return
		}
		var req struct {
			Text       string `json:"text"`
			SourceLang string `json:"source_lang"`
			TargetLang string `json:"target_lang"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
			return
		}
		if req.SourceLang == "" {
			req.SourceLang = "auto-detect"
		}
		if req.TargetLang == "" {
			req.TargetLang = "English"
		}
		result, err := gemini.TranslateMessage(c.Request.Context(), req.Text, req.SourceLang, req.TargetLang)
		if err != nil {
			log.Printf("AI translate error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "translation failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"translation": result})
	}
}

// 7. AI Progress Report
func handleProgressReport(gemini *ai.GeminiClient, repo *repository.FirestoreRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureAIAvailable(c, gemini) {
			return
		}
		reports, err := repo.GetAllReports(c.Request.Context(), 100)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reports"})
			return
		}
		result, err := gemini.GenerateProgressReport(c.Request.Context(), reports)
		if err != nil {
			log.Printf("AI progress report error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "progress report generation failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"progress_report": result})
	}
}

// 8. Volunteer Skill Recommendations
func handleRecommendSkills(gemini *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureAIAvailable(c, gemini) {
			return
		}
		var req struct {
			CurrentSkills      []string `json:"current_skills"`
			CompletedTaskTypes []string `json:"completed_task_types"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		result, err := gemini.RecommendSkills(c.Request.Context(), req.CurrentSkills, req.CompletedTaskTypes)
		if err != nil {
			log.Printf("AI skill recommendation error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "skill recommendation failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"recommendations": result})
	}
}

// 9. Document OCR
func handleOCR(gemini *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureAIAvailable(c, gemini) {
			return
		}
		var req struct {
			Image string `json:"image"` // base64
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Image == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image (base64) is required"})
			return
		}
		result, err := gemini.OCRDocument(c.Request.Context(), req.Image)
		if err != nil {
			log.Printf("AI OCR error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "OCR failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ocr_result": result})
	}
}

// 10. AI Chatbot
func handleAIChat(gemini *ai.GeminiClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ensureAIAvailable(c, gemini) {
			return
		}
		var req struct {
			Message     string `json:"message"`
			TaskContext string `json:"task_context"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Message == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
			return
		}
		response, err := gemini.ChatWithAI(c.Request.Context(), req.Message, req.TaskContext)
		if err != nil {
			log.Printf("AI chat error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "chat failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"response": response})
	}
}
