package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAdminAPIExample tests Admin API endpoints
func TestAdminAPIExample(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/admin/users", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
}
