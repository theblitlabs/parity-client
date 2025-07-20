package types

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, statusCode int, message string) error {
	return WriteJSON(w, statusCode, HTTPError{
		Status:  statusCode,
		Message: message,
	})
}

func CopyHeaders(dst, src http.Header) {
	for header, values := range src {
		for _, value := range values {
			dst.Add(header, value)
		}
	}
}

func ReadJSONBody(body io.ReadCloser, v interface{}) error {
	defer func() {
		if closeErr := body.Close(); closeErr != nil {
			log.Printf("Error closing body: %v", closeErr)
		}
	}()
	return json.NewDecoder(body).Decode(v)
}

func CopyBody(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}
