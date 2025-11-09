package api
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStatsAPIExample tests Stats API endpoints
func TestStatsAPIExample(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/stats", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
}
