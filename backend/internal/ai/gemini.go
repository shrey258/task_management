package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

type TaskSuggestion struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Tags        []string `json:"tags"`
	SubTasks    []string `json:"sub_tasks,omitempty"`
}

func NewGeminiService() (*GeminiService, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	// Create a generative model
	model := client.GenerativeModel("gemini-pro")

	return &GeminiService{
		client: client,
		model:  model,
	}, nil
}

func (s *GeminiService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *GeminiService) GenerateTaskSuggestions(ctx context.Context, description string) ([]TaskSuggestion, error) {
	prompt := fmt.Sprintf(`Based on the following task description, suggest a breakdown of tasks with priorities and tags.
Description: %s

Please provide the response in the following JSON format:
[
  {
    "title": "Main task title",
    "description": "Detailed description",
    "priority": "high|medium|low",
    "tags": ["tag1", "tag2"],
    "sub_tasks": ["Sub-task 1", "Sub-task 2"]
  }
]

Focus on actionable items and clear priorities. Keep descriptions concise but informative.`, description)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no suggestions generated")
	}

	// Extract the JSON response
	jsonStr := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	var suggestions []TaskSuggestion
	if err := json.Unmarshal([]byte(jsonStr), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse suggestions: %v", err)
	}

	return suggestions, nil
}

func (s *GeminiService) AnalyzeTaskPriority(ctx context.Context, title, description string) (*TaskSuggestion, error) {
	prompt := fmt.Sprintf(`Analyze the following task and suggest appropriate priority and tags:
Title: %s
Description: %s

Please provide the response in the following JSON format:
{
  "title": "Original title",
  "description": "Original description with any suggested improvements",
  "priority": "high|medium|low",
  "tags": ["tag1", "tag2"]
}

Consider factors like urgency, impact, and complexity when determining priority.`, title, description)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to analyze task: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no analysis generated")
	}

	// Extract the JSON response
	jsonStr := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	var suggestion TaskSuggestion
	if err := json.Unmarshal([]byte(jsonStr), &suggestion); err != nil {
		return nil, fmt.Errorf("failed to parse analysis: %v", err)
	}

	return &suggestion, nil
}
