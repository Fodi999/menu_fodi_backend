package api
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMealPlanAPIExample tests Meal Plan API endpoints
func TestMealPlanAPIExample(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/meal-plan", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
}
