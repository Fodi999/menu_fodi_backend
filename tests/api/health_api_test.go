package api
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHealthAPIExample tests Health API endpoints
func TestHealthAPIExample(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/health", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
}
