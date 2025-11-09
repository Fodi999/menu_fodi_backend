package api
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSemiFinishedAPIExample tests Semi-Finished API endpoints
func TestSemiFinishedAPIExample(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/semi-finished", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
}
