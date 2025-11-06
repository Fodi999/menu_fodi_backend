package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/ai"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// ChefMentorStreamRequest for streaming endpoint
type ChefMentorStreamRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"sessionId,omitempty"`
	Language  string `json:"language,omitempty"`
}

// ChefMentorStreamHandler streams AI responses in real-time
// POST /api/ai/chef-mentor/stream
func ChefMentorStreamHandler(w http.ResponseWriter, r *http.Request) {
	var req ChefMentorStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	repo := database.NewChefMentorRepository()

	// Get or create session
	var dbSession *models.ChefMentorSession
	var err error

	if req.SessionID != "" {
		dbSession, err = repo.GetSession(req.SessionID)
		if err != nil {
			fmt.Fprintf(w, "event: error\ndata: {\"error\":\"Session not found\"}\n\n")
			flusher.Flush()
			return
		}
	} else {
		lang := req.Language
		if lang == "" {
			lang = "ua"
		}
		dbSession, err = repo.CreateSession(nil, lang)
		if err != nil {
			fmt.Fprintf(w, "event: error\ndata: {\"error\":\"Failed to create session\"}\n\n")
			flusher.Flush()
			return
		}
	}

	sessionIDStr := dbSession.ID.String()

	// Send session ID first
	fmt.Fprintf(w, "event: session\ndata: {\"sessionId\":\"%s\"}\n\n", sessionIDStr)
	flusher.Flush()

	// Get conversation history
	dbMessages, err := repo.GetMessages(sessionIDStr)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\":\"Failed to load messages\"}\n\n")
		flusher.Flush()
		return
	}

	// Parse current recipe
	currentRecipe := &RecipeDraft{}
	if len(dbSession.Recipe) > 0 {
		recipeJSON, _ := json.Marshal(dbSession.Recipe)
		json.Unmarshal(recipeJSON, currentRecipe)
	}

	// Build system prompt
	systemPrompt := buildMentorSystemPrompt(dbSession.Language)
	contextPrompt := buildRecipeContext(currentRecipe, dbSession.Language)

	// Create AI messages
	client := ai.NewGroqClient()
	messages := []ai.GroqMessage{
		{
			Role:    "system",
			Content: systemPrompt + "\n\n" + contextPrompt,
		},
	}

	// Add history
	for _, msg := range dbMessages {
		messages = append(messages, ai.GroqMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current user message
	messages = append(messages, ai.GroqMessage{
		Role:    "user",
		Content: req.Message,
	})

	// Save user message to DB
	repo.SaveMessage(dbSession.ID, "user", req.Message)

	// Stream AI response
	stream, err := client.CreateChatCompletionStream(messages, 0.7, 1000)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\":\"AI service error\"}\n\n")
		flusher.Flush()
		return
	}
	defer stream.Close()

	var fullResponse strings.Builder

	// Stream tokens
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(w, "event: error\ndata: {\"error\":\"Stream error\"}\n\n")
			flusher.Flush()
			return
		}

		content := response.Choices[0].Delta.Content
		fullResponse.WriteString(content)

		// Send token to client
		contentJSON, _ := json.Marshal(map[string]string{"content": content})
		fmt.Fprintf(w, "event: token\ndata: %s\n\n", contentJSON)
		flusher.Flush()
	}

	assistantMessage := fullResponse.String()

	// Save assistant message to DB
	repo.SaveMessage(dbSession.ID, "assistant", assistantMessage)

	// Extract recipe updates
	currentRecipe = smartExtractRecipeUpdates(assistantMessage, currentRecipe, req.Message, dbSession.Language)

	// Update session
	repo.UpdateSession(sessionIDStr, currentRecipe, nil)

	// Determine completion
	isComplete := isRecipeComplete(currentRecipe)
	if isComplete {
		repo.MarkComplete(sessionIDStr)
	}

	// Generate quick replies
	quickReplies := generateQuickReplies(currentRecipe, dbSession.Language)

	// Send final metadata
	metadata := map[string]interface{}{
		"recipe":       currentRecipe,
		"isComplete":   isComplete,
		"quickReplies": quickReplies,
	}
	metadataJSON, _ := json.Marshal(metadata)
	fmt.Fprintf(w, "event: metadata\ndata: %s\n\n", metadataJSON)
	flusher.Flush()

	// Send done event
	fmt.Fprintf(w, "event: done\ndata: {\"status\":\"success\"}\n\n")
	flusher.Flush()
}
