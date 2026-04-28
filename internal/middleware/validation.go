package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ValidateReportInput validates report creation request
func ValidateReportInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RawText     string  `json:"raw_text"`
			IssueType   string  `json:"issue_type"`
			UserUrgency string  `json:"user_urgency"`
			Latitude    float64 `json:"latitude"`
			Longitude   float64 `json:"longitude"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("validation error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
			c.Abort()
			return
		}

		// Validate required fields
		if req.RawText == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "raw_text is required"})
			c.Abort()
			return
		}

		if req.IssueType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "issue_type is required"})
			c.Abort()
			return
		}

		if req.UserUrgency == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_urgency is required"})
			c.Abort()
			return
		}

		// Validate coordinates
		if req.Latitude < -90 || req.Latitude > 90 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "latitude must be between -90 and 90"})
			c.Abort()
			return
		}

		if req.Longitude < -180 || req.Longitude > 180 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "longitude must be between -180 and 180"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateVolunteerInput validates volunteer registration request
func ValidateVolunteerInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name      string   `json:"name"`
			Skills    []string `json:"skills"`
			Latitude  float64  `json:"latitude"`
			Longitude float64  `json:"longitude"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("validation error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
			c.Abort()
			return
		}

		// Validate required fields
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			c.Abort()
			return
		}

		if len(req.Skills) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one skill is required"})
			c.Abort()
			return
		}

		// Validate coordinates
		if req.Latitude < -90 || req.Latitude > 90 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "latitude must be between -90 and 90"})
			c.Abort()
			return
		}

		if req.Longitude < -180 || req.Longitude > 180 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "longitude must be between -180 and 180"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateCaseInput validates case creation request
func ValidateCaseInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Title           string   `json:"title"`
			LinkedReportIDs []string `json:"linked_report_ids"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("validation error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
			c.Abort()
			return
		}

		// Validate required fields
		if req.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
			c.Abort()
			return
		}

		if len(req.LinkedReportIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one linked_report_id is required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
