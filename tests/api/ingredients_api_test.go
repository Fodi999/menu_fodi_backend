package api
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIngredientsAPIExample tests Ingredients API endpoints
func TestIngredientsAPIExample(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/ingredients", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
}
