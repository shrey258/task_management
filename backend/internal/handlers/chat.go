package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ChatHandler struct {
	model *genai.GenerativeModel
}

func NewChatHandler() (*ChatHandler, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	log.Printf("Initializing Gemini client with API key length: %d", len(apiKey))

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Failed to create Gemini client: %v", err)
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	model := client.GenerativeModel("gemini-pro")
	
	// Test the model with a simple prompt
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	testResp, err := model.GenerateContent(testCtx, genai.Text("Hello"))
	if err != nil {
		log.Printf("Failed to test Gemini model: %v", err)
		return nil, fmt.Errorf("failed to test Gemini model: %v", err)
	}
	
	if len(testResp.Candidates) == 0 {
		log.Printf("Gemini model test returned no candidates")
		return nil, fmt.Errorf("Gemini model test returned no candidates")
	}

	log.Printf("Successfully initialized Gemini client and tested model")
	return &ChatHandler{model: model}, nil
}

type ChatRequest struct {
	Message string `json:"message"`
	Type    string `json:"type"` // "general" or "breakdown"
}

type ChatResponse struct {
	Response string   `json:"response"`
	Tasks    []Task   `json:"tasks,omitempty"`
}

type Task struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Subtasks    []string `json:"subtasks,omitempty"`
}

func (h *ChatHandler) HandleChat(c *fiber.Ctx) error {
	startTime := time.Now()
	log.Printf("Received chat request from user: %v", c.Locals("user_id"))

	// Check if model is initialized
	if h.model == nil {
		log.Println("Chat handler error: model not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "AI service not available",
		})
	}

	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Chat handler error: invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Message == "" {
		log.Println("Chat handler error: empty message received")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "message cannot be empty",
		})
	}

	log.Printf("Processing chat request with message length: %d", len(req.Message))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var prompt string
	if req.Type == "breakdown" {
		prompt = fmt.Sprintf(`Break down the following task into smaller, manageable subtasks. For each subtask, provide a clear description and suggest a priority level (high, medium, or low). Format the response as a structured task breakdown.

Task: %s

Please provide the breakdown in this format:
Main Task Title: [Title]
Description: [Overall task description]
Priority: [Priority level]
Subtasks:
1. [Subtask 1 description]
2. [Subtask 2 description]
3. [Subtask 3 description]
...`, req.Message)
	} else {
		prompt = fmt.Sprintf(`You are a task management AI assistant. Help the user with their question or task: %s

Provide clear, actionable advice and suggestions for better task management.`, req.Message)
	}

	log.Printf("Sending prompt to Gemini: %s", prompt)
	resp, err := h.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Chat handler error: failed to generate response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to generate AI response: %v", err),
		})
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		log.Printf("Chat handler error: no candidates in response")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "AI returned no response",
		})
	}

	response := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			response += string(textPart)
		}
	}

	log.Printf("Generated response length: %d", len(response))
	log.Printf("Received response from Gemini API in %v", time.Since(startTime))
	return c.JSON(ChatResponse{
		Response: response,
	})
}
