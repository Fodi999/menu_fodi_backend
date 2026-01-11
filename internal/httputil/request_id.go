package httputil

import (
	"context"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
)

const RequestIDHeader = "X-Request-ID"

// AddRequestIDToHeaders adds request_id from context to HTTP headers
// Use this when making downstream HTTP calls (AI services, external APIs, etc.)
//
// Example:
//
//	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat", body)
//	httputil.AddRequestIDToHeaders(ctx, req.Header)
//	resp, err := client.Do(req)
func AddRequestIDToHeaders(ctx context.Context, headers http.Header) {
	requestID := logger.GetRequestID(ctx)
	if requestID != "" && requestID != "unknown" {
		headers.Set(RequestIDHeader, requestID)
	}
}

// PropagateRequestID is a helper to propagate request_id to downstream services
// Returns the request_id string for manual header setting
//
// Example (Groq API call):
//
//	requestID := httputil.PropagateRequestID(ctx)
//	req.Header.Set("X-Request-ID", requestID)
//	req.Header.Set("X-Correlation-ID", requestID) // Some services use this
func PropagateRequestID(ctx context.Context) string {
	return logger.GetRequestID(ctx)
}
