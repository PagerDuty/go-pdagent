package common

// Auto-injected at build time.
var GitCommit string

// Manually updated, should correspond to Git release tags.
const Version = "0.0.1"

func init() {
	if GitCommit == "" {
		GitCommit = "unavailable"
	}
}
