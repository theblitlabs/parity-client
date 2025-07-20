package client

import (
	"fmt"
	"net"
)

func IsPortAvailable(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("port %d is not available: %v", port, err)
	}
	if err := ln.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %v", err)
	}
	return nil
}
