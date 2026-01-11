package models

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
)

// APIResponse is the standard response format for all endpoints
type APIResponse struct {
	Data  interface{}    `json:"data"`
	Error *APIError      `json:"error"`
	Meta  *ResponseMeta  `json:"meta"`
}

// APIError represents an error in the API
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ResponseMeta contains metadata about the response
type ResponseMeta struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version,omitempty"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedData wraps data with pagination info
type PaginatedData struct {
	Items      interface{}     `json:"items"`
	Pagination *PaginationInfo `json:"pagination"`
}

// SuccessResponse sends a successful response with 200 OK
func SuccessResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	SuccessResponseWithStatus(w, r, http.StatusOK, data)
}

// CreatedResponse sends a successful response with 201 Created
func CreatedResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	SuccessResponseWithStatus(w, r, http.StatusCreated, data)
}

// SuccessResponseWithStatus sends a successful response with custom status
func SuccessResponseWithStatus(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	response := APIResponse{
		Data:  data,
		Error: nil,
		Meta:  buildMeta(r.Context()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// PaginatedResponse sends a paginated response
func PaginatedResponse(w http.ResponseWriter, r *http.Request, items interface{}, page, limit, total int) {
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	data := PaginatedData{
		Items: items,
		Pagination: &PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	SuccessResponse(w, r, data)
}

// ErrorResponse sends an error response
func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, code, message, details string) {
	response := APIResponse{
		Data: nil,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: buildMeta(r.Context()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// BadRequestError sends a 400 Bad Request error
func BadRequestError(w http.ResponseWriter, r *http.Request, code, message, details string) {
	ErrorResponse(w, r, http.StatusBadRequest, code, message, details)
}

// UnauthorizedError sends a 401 Unauthorized error
func UnauthorizedError(w http.ResponseWriter, r *http.Request, code, message, details string) {
	ErrorResponse(w, r, http.StatusUnauthorized, code, message, details)
}

// ForbiddenError sends a 403 Forbidden error
func ForbiddenError(w http.ResponseWriter, r *http.Request, code, message, details string) {
	ErrorResponse(w, r, http.StatusForbidden, code, message, details)
}

// NotFoundError sends a 404 Not Found error
func NotFoundError(w http.ResponseWriter, r *http.Request, code, message, details string) {
	ErrorResponse(w, r, http.StatusNotFound, code, message, details)
}

// InternalServerError sends a 500 Internal Server Error
func InternalServerError(w http.ResponseWriter, r *http.Request, code, message, details string) {
	ErrorResponse(w, r, http.StatusInternalServerError, code, message, details)
}

// buildMeta creates response metadata
func buildMeta(ctx context.Context) *ResponseMeta {
	requestID := logger.GetRequestID(ctx)

	return &ResponseMeta{
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "v1",
	}
}
