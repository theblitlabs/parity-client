package version

import (
	"fmt"
	"runtime"
)

var (
	// These will be set during build
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

// GetVersion returns the full version information
func GetVersion() string {
	return fmt.Sprintf("v%s", Version)
}

// GetBuildInfo returns detailed build information
func GetBuildInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"commit":     CommitSHA,
		"build_time": BuildTime,
		"go_version": GoVersion,
		"platform":   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetShortVersion returns just the version string
func GetShortVersion() string {
	return Version
}
