package middleware

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger is a middleware that logs HTTP requests
func Logger() fiber.Handler {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
	}

	// Open log file
	logFile, err := os.OpenFile("logs/server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
	}

	// Create multi-writer to write to both file and stdout
	log.SetOutput(logFile)

	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		// Get request body if it exists
		var body string
		if method == "POST" || method == "PUT" {
			body = string(c.Body())
			if len(body) > 1000 { // Truncate long bodies
				body = body[:1000] + "..."
			}
		}

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get response status
		status := c.Response().StatusCode()

		// Get user ID from locals if authenticated
		userID := c.Locals("user_id")
		if userID == nil {
			userID = "anonymous"
		}

		// Format log entry
		logEntry := fmt.Sprintf(
			"[%s] %s | %s %s | Status: %d | Duration: %v | UserID: %v",
			time.Now().Format("2006-01-02 15:04:05"),
			c.IP(),
			method,
			path,
			status,
			duration,
			userID,
		)

		// Add request body for POST/PUT requests
		if body != "" {
			logEntry += fmt.Sprintf(" | Body: %s", body)
		}

		// Add error if any
		if err != nil {
			logEntry += fmt.Sprintf(" | Error: %v", err)
		}

		// Write to both console and file
		log.Println(logEntry)
		fmt.Println(logEntry)

		return err
	}
}
