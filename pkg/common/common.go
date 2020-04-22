package common

import (
	"os"
	"time"
)

// Normally set at build time.
var Commit = "unavailable"
var Date = time.Now().Format(time.RFC3339)
var Version = "unavailable"

func init() {
	if Commit == "" {
		Commit = "unavailable"
	}

	if Date == "" {
		Date = time.Now().Format(time.RFC3339)
	}

	if Version == "" {
		Version = "unavailable"
	}
}

func IsProduction() bool {
	return os.Getenv("APP_ENV") == "production"
}
