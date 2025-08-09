package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateRequest validates the request body against the provided struct
func ValidateRequest(next http.HandlerFunc, data interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode the request body into the provided struct
		if err := json.NewDecoder(r.Body).Decode(data); err != nil {
			handleError(w, err)
			return
		}
		defer r.Body.Close()

		// Validate the struct
		if err := validate.Struct(data); err != nil {
			handleError(w, err)
			return
		}

		// Store the validated data in the request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "validatedData", data)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}
}

// GetValidatedData retrieves the validated data from the request context
func GetValidatedData(r *http.Request, data interface{}) bool {
	if val := r.Context().Value("validatedData"); val != nil {
		// Use JSON marshaling/unmarshaling to copy the data
		// This is a simple way to handle the copy generically
		bytes, err := json.Marshal(val)
		if err != nil {
			return false
		}
		if err := json.Unmarshal(bytes, data); err != nil {
			return false
		}
		return true
	}
	return false
}
