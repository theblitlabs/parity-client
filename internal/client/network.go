package client

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net"
)

// IsPortAvailable checks if a port is available for use
func IsPortAvailable(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("port %d is not available: %v", port, err)
	}
	ln.Close()
	return nil
}

// NewMultipartWriter creates a new multipart writer
func NewMultipartWriter(body *bytes.Buffer) *multipart.Writer {
	writer := multipart.NewWriter(body)
	return writer
}

// AddFileToWriter adds a file to a multipart writer
func AddFileToWriter(writer *multipart.Writer, fieldName, filename string, reader io.Reader) error {
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	_, err = io.Copy(part, reader)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	return nil
}
