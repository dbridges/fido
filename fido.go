package fido

import (
	"encoding/json"
	"net/http"
)

// H is an alias for map[string]any, useful for quickly generating JSON objects
type H map[string]any

// JSON encodes a value and writes it to the supplied response writer after
// setting the response status.
func JSON(w http.ResponseWriter, status int, d any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(d)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Error writing JSON")
	}
}

// JSONError wraps the message in a JSON object and writes it to the supplied
// response writer
func JSONError(w http.ResponseWriter, status int, message string) {
	JSON(w, status, H{"error": message})
}

// BindJSON decodes the http request body into v
func BindJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
