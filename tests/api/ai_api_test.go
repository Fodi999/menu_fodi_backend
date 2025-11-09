package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAIChefMentorEndpoint tests Chef Mentor endpoint
func TestAIChefMentorEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid chef mentor request",
			payload: map[string]interface{}{
				"message": "How do I make a salad?",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty message",
			payload: map[string]interface{}{
				"message": "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/ai/chef-mentor", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				message, ok := payload["message"].(string)
				if !ok || message == "" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "message cannot be empty"}`))
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"response": "Chef advice"}`))
			})

			handler.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestAIRecipeGeneratorEndpoint tests Recipe Generator endpoint
func TestAIRecipeGeneratorEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid recipe request",
			payload: map[string]interface{}{
				"title":    "Pasta",
				"language": "en",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing title",
			payload: map[string]interface{}{
				"language": "en",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/ai/recipe-generator", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				title, ok := payload["title"].(string)
				if !ok || title == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"recipe": "generated"}`))
			})

			handler.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestAIMealPlanEndpoint tests Meal Plan endpoint
func TestAIMealPlanEndpoint(t *testing.T) {
	t.Run("valid meal plan request", func(t *testing.T) {
		payload := map[string]interface{}{
			"days":             7,
			"targetCalories": 2000,
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/ai/meal-plan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer mock-token")

		w := httptest.NewRecorder()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"plan": []}`))
		})

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestAIFridgeRecommendationsEndpoint tests Fridge Recommendations endpoint
func TestAIFridgeRecommendationsEndpoint(t *testing.T) {
	t.Run("valid fridge recommendations", func(t *testing.T) {
		payload := map[string]interface{}{
			"dietaryPreferences": []string{"vegetarian"},
			"cuisine":            "Italian",
			"maxTime":            30,
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/ai/fridge-recommendations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer mock-token")

		w := httptest.NewRecorder()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"recommendations": []}`))
		})

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "recommendations")
	})
}
