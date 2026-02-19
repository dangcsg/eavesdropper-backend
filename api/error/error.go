package error

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Code    int    `json:"code"`              // http status code
	ID      string `json:"id"`                // Identifies the error for specific handling. Empty for unhandled errors in client side.
	Message string `json:"message,omitempty"` // Message for dev
}

func WriteJSONError(w http.ResponseWriter, status int, id, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(
		APIError{
			Code:    status,
			ID:      id,
			Message: msg,
		})
}
