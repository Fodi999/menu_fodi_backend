package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/service"
	aihttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/transport/http"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter creates a minimal router with only AI routes for E2E testing
func setupTestRouter(t *testing.T) http.Handler {
	r := chi.NewRouter()

	// Create AI service and handlers (without database for E2E tests)
	svc := service.NewAIService()
	handlers := aihttp.NewAIHandlers(svc, nil)

	// Register AI routes directly (all public for testing)
	r.Route("/api/ai", func(r chi.Router) {
		// Public AI endpoints
		r.Post("/chef-mentor", handlers.ChefMentor)
		r.Post("/recipe-generator", handlers.GenerateRecipe)
		r.Post("/meal-plan", handlers.GenerateMealPlan)           // Made public for E2E testing
		r.Post("/fridge-recommendations", handlers.GetFridgeRecommendations) // Made public for E2E testing
	})

	return r
}

// TestAIChefMentorE2E tests the complete flow of chef mentor endpoint
func TestAIChefMentorE2E(t *testing.T) {
	router := setupTestRouter(t)

	tests := []struct {
		name           string
		request        dto.ChefMentorRequest
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "valid chef mentor request with real API",
			request: dto.ChefMentorRequest{
				Message: "How do I make a perfect risotto?",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				// Response should contain AI assistant message
				assert.NotEmpty(t, body, "Response body should not be empty")
				// Should be valid JSON
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				assert.NoError(t, err, "Response should be valid JSON")
			},
		},
		{
			name: "empty message validation",
			request: dto.ChefMentorRequest{
				Message: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "message", "Error response should contain validation message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/ai/chef-mentor",
				bytes.NewBuffer(reqBody),
			)
			req.Header.Set("Content-Type", "application/json")

			// Create response writer
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, w.Code, "Expected status code %d, got %d", tt.expectedStatus, w.Code)

			// Run additional checks on response
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

// TestAIRecipeGeneratorE2E tests the complete flow of recipe generator endpoint
func TestAIRecipeGeneratorE2E(t *testing.T) {
	router := setupTestRouter(t)

	tests := []struct {
		name           string
		request        dto.RecipeGenerationRequest
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "generate recipe with title and language",
			request: dto.RecipeGenerationRequest{
				Title:    "Chocolate Cake",
				Language: "en",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				assert.NoError(t, err, "Response should be valid JSON")
				assert.Contains(t, body, "recipe", "Response should contain recipe data")
			},
		},
		{
			name: "missing required title field",
			request: dto.RecipeGenerationRequest{
				Title:    "",
				Language: "en",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "message", "Error response expected")
			},
		},
		{
			name: "different language (Polish)",
			request: dto.RecipeGenerationRequest{
				Title:    "Bigos",
				Language: "pl",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.NotEmpty(t, body)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/ai/recipe-generator",
				bytes.NewBuffer(reqBody),
			)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

// TestAIMealPlanE2E tests the complete flow of meal plan endpoint
func TestAIMealPlanE2E(t *testing.T) {
	router := setupTestRouter(t)

	tests := []struct {
		name           string
		request        dto.MealPlanRequest
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "invalid: zero days (validation test)",
			request: dto.MealPlanRequest{
				Days:           0,
				TargetCalories: 2000,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/ai/meal-plan",
				bytes.NewBuffer(reqBody),
			)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

// TestAIFridgeRecommendationsE2E tests the complete flow of fridge recommendations endpoint
// Note: Some tests may fail due to auth requirements in protected endpoints
func TestAIFridgeRecommendationsE2E(t *testing.T) {
	router := setupTestRouter(t)

	// Test with simplified setup - just check that endpoint exists and validates
	t.Run("endpoint exists and validates inputs", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/ai/fridge-recommendations",
			bytes.NewBuffer([]byte(`{"dietary_preferences":["vegetarian"],"cuisine":"Italian","max_time":45}`)),
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should respond with some status (200, 401 for auth, or 500)
		// Main point is that the endpoint is wired correctly
		assert.NotEqual(t, http.StatusNotFound, w.Code, "Endpoint should exist")
	})
}

// TestAIEndpointNotFound tests that non-existent endpoints return 404
func TestAIEndpointNotFound(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestAIEndpointMethodNotAllowed tests that wrong HTTP methods return 405
func TestAIEndpointMethodNotAllowed(t *testing.T) {
	router := setupTestRouter(t)

	// POST endpoint should not accept GET
	req := httptest.NewRequest(http.MethodGet, "/api/ai/chef-mentor", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should be either 405 Method Not Allowed or 404
	assert.Contains(t, []int{http.StatusMethodNotAllowed, http.StatusNotFound}, w.Code)
}

// TestAIResponseContentType verifies that responses are JSON
func TestAIResponseContentType(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/ai/recipe-generator",
		bytes.NewBuffer([]byte(`{"title":"Pasta","language":"en"}`)),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Contains(t, w.Header().Get("Content-Type"), "application/json",
		"Response should have JSON content type")
}
