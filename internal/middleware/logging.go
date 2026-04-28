package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs all incoming HTTP requests with status, method, and duration
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		logLevel := "INFO"
		if statusCode >= 400 && statusCode < 500 {
			logLevel = "WARN"
		} else if statusCode >= 500 {
			logLevel = "ERROR"
		}

		log.Printf("[%s] %s %s - %d (%dms)", logLevel, method, path, statusCode, duration.Milliseconds())
	}
}

// ErrorHandler returns a middleware that normalizes error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("[ERROR] %s: %v", c.Request.URL.Path, err)
			}

			// If no response status was set, return 500
			if c.Writer.Status() == http.StatusOK {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}
	}
}
