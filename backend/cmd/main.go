package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
	"github.com/shrey258/task_management/internal/ai"
	"github.com/shrey258/task_management/internal/database"
	"github.com/shrey258/task_management/internal/handlers"
	"github.com/shrey258/task_management/internal/middleware"
	"github.com/shrey258/task_management/internal/repository"
	ws "github.com/shrey258/task_management/internal/websocket"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	// Connect to MongoDB
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Close()

	// Initialize Gemini service
	gemini, err := ai.NewGeminiService()
	if err != nil {
		log.Fatalf("Failed to initialize Gemini service: %v", err)
	}
	defer gemini.Close()

	// Create Fiber app with custom config
	app := fiber.New(fiber.Config{
		AppName: "Task Management API",
	})

	// Add logger middleware
	app.Use(logger.New())

	// Configure CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("FRONTEND_URL"),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"message": "Server is running",
		})
	})

	// Setup routes
	setupRoutes(app, hub, gemini)

	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Log startup information
	log.Printf("Starting server on port %s", port)
	log.Printf("Environment: %s", os.Getenv("GO_ENV"))
	log.Printf("MongoDB URI: %s", os.Getenv("MONGODB_URI"))

	// Start server with explicit host and port
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func setupRoutes(app *fiber.App, hub *ws.Hub, gemini *ai.GeminiService) {
	// Initialize repositories
	userRepo := repository.NewUserRepository()
	taskRepo := repository.NewTaskRepository()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo)
	taskHandler := handlers.NewTaskHandler(taskRepo, hub)
	wsHandler := handlers.NewWebSocketHandler(hub)
	aiHandler := handlers.NewAIHandler(gemini)
	log.Println("Initializing chat handler...")
	chatHandler, err := handlers.NewChatHandler()
	if err != nil {
		log.Printf("Warning: Failed to initialize chat handler: %v", err)
	}

	// Auth routes
	auth := app.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected routes
	protected := app.Group("/api", middleware.Protected())
	
	// User routes
	protected.Get("/user", authHandler.GetCurrentUser)

	// Task routes
	tasks := protected.Group("/tasks")
	tasks.Post("/", taskHandler.CreateTask)
	tasks.Get("/", taskHandler.GetTasks)
	tasks.Get("/:id", taskHandler.GetTask)
	tasks.Put("/:id", taskHandler.UpdateTask)
	tasks.Delete("/:id", taskHandler.DeleteTask)

	// AI routes
	ai := protected.Group("/ai")
	ai.Post("/suggest", aiHandler.GenerateTaskSuggestions)
	ai.Post("/analyze", aiHandler.AnalyzeTask)

	// Chat route
	if chatHandler != nil {
		protected.Post("/chat", chatHandler.HandleChat)
	}

	// WebSocket route
	app.Use("/ws", wsHandler.UpgradeConnection)
	app.Get("/ws", websocket.New(wsHandler.HandleWebSocket, websocket.Config{
		Filter: func(c *fiber.Ctx) bool {
			return true // You can add additional filtering here
		},
	}))
}
