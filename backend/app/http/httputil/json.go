package httputil

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response with the given status code
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Log error but can't change status at this point
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

// Error writes a JSON error response
func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, map[string]string{"error": message})
}

// DecodeJSON decodes JSON from the request body into the target
func DecodeJSON(r *http.Request, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}

// Data writes raw data with the given content type
func Data(w http.ResponseWriter, statusCode int, contentType string, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	_, _ = w.Write(data)
}
