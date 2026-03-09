package handlers

import (
	"fmt"
	"strings"
)

func normalizeServerBaseURL(serverURL string) string {
	baseURL := strings.TrimRight(serverURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/api")
	return baseURL
}

func localProxyBaseURL(host string, port int) string {
	displayHost := host
	switch displayHost {
	case "", "0.0.0.0", "::":
		displayHost = "127.0.0.1"
	}

	return fmt.Sprintf("http://%s:%d", displayHost, port)
}

func NormalizeServerBaseURL(serverURL string) string {
	return normalizeServerBaseURL(serverURL)
}

func LocalProxyBaseURL(host string, port int) string {
	return localProxyBaseURL(host, port)
}
