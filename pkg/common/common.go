package common

import "time"

// Normally set at build time.
var Commit = "unavailable"
var Date = time.Now().Format(time.RFC3339)
var Version = "unavailable"
