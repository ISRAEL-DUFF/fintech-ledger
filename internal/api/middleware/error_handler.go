package middleware

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// APIError represents a standardized API error response
type APIError struct {
	Status  int         `json:"-"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new API error
func NewAPIError(status int, message string, details interface{}) *APIError {
	return &APIError{
		Status:  status,
		Message: message,
		Details: details,
	}
}

// ErrorHandler is a middleware that handles errors and returns JSON responses
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = errors.New("unknown error occurred")
				}
				handleError(w, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// handleError processes errors and writes the appropriate response
func handleError(w http.ResponseWriter, err error) {
	// Default to 500 Internal Server Error
	statusCode := http.StatusInternalServerError
	var details interface{}

	switch e := err.(type) {
	case *APIError:
		statusCode = e.Status
		details = e.Details
	case validator.ValidationErrors:
		statusCode = http.StatusBadRequest
		errMsgs := make(map[string]string)
		for _, ve := range e {
			errMsgs[ve.Field()] = ve.Tag()
		}
		details = errMsgs
		err = errors.New("validation failed")
	case *json.SyntaxError, *json.UnmarshalTypeError:
		statusCode = http.StatusBadRequest
		err = errors.New("invalid JSON payload")
	}

	// Log the error for server-side debugging
	if statusCode >= 500 {
		log.Printf("Internal server error: %v", err)
	}

	// Write the error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResponse := map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": err.Error(),
	}

	if details != nil {
		errResponse["details"] = details
	}

	if err := json.NewEncoder(w).Encode(errResponse); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// JSONMiddleware sets the Content-Type header to application/json
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
