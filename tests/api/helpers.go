package api
package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================
// Test Helpers for API Tests
// ============================================

// HTTPTestHelper wraps httptest functionality
type HTTPTestHelper struct {
	t *testing.T
}

// NewHTTPTestHelper creates a new HTTP test helper
func NewHTTPTestHelper(t *testing.T) *HTTPTestHelper {
	return &HTTPTestHelper{t: t}
}

// MakeRequest creates and executes an HTTP request
func (h *HTTPTestHelper) MakeRequest(method, path string, body io.Reader) *httptest.ResponseRecorder {
	h.t.Helper()

	req, err := http.NewRequest(method, path, body)
	require.NoError(h.t, err)

	w := httptest.NewRecorder()
	return w
}

// AssertStatusCode asserts the HTTP status code
func (h *HTTPTestHelper) AssertStatusCode(w *httptest.ResponseRecorder, expected int) {
	h.t.Helper()
	if w.Code != expected {
		h.t.Errorf("Expected status code %d, got %d", expected, w.Code)
	}
}

// AssertContentType asserts the Content-Type header
func (h *HTTPTestHelper) AssertContentType(w *httptest.ResponseRecorder, expected string) {
	h.t.Helper()
	actual := w.Header().Get("Content-Type")
	if actual != expected {
		h.t.Errorf("Expected Content-Type %s, got %s", expected, actual)
	}
}

// GetResponseBody returns the response body as string
func (h *HTTPTestHelper) GetResponseBody(w *httptest.ResponseRecorder) string {
	h.t.Helper()
	return w.Body.String()
}

// MockResponseHandler creates a mock HTTP handler
func MockResponseHandler(statusCode int, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	})
}

// TestServer creates a test HTTP server
func TestServer(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}
