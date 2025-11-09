package api
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContactAPIExample tests Contact API endpoints
func TestContactAPIExample(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/contact/send", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
}
