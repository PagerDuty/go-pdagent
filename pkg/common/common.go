package common

import "time"

// Commit normally auto-injected at build time.
var Commit = ""

// Date normally auto-injected at build time.
var Date = ""

// Version normally auto-injected at build time.
var Version = ""

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
