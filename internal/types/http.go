package types

import (
	"encoding/json"
	"io"
	"net/http"
)

// ResponseWriter wraps common response writing operations
type ResponseWriter struct {
	http.ResponseWriter
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response with the given status code
func WriteError(w http.ResponseWriter, statusCode int, message string) error {
	return WriteJSON(w, statusCode, HTTPError{
		Status:  statusCode,
		Message: message,
	})
}

// CopyHeaders copies headers from source to destination
func CopyHeaders(dst, src http.Header) {
	for header, values := range src {
		for _, value := range values {
			dst.Add(header, value)
		}
	}
}

// ReadJSONBody reads and decodes a JSON request body into the given struct
func ReadJSONBody(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(v)
}

// CopyBody copies the body from src to dst
func CopyBody(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}
