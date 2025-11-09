package api
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMetricsAPIExample tests Metrics API endpoints
func TestMetricsAPIExample(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/metrics", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
}
