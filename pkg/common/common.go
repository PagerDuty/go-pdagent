package common

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/spf13/viper"
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

func UserAgent() string {
	version := Version
	system := runtime.GOOS
	commit := Commit
	date := Date

	return fmt.Sprintf("go-pdagent/%v (%v, commit: %v, date: %v)", version, system, commit, date)
}

func PdEventsUrl() string {
	region := viper.GetString("region")
	if region == "eu" {
		return "https://events.eu.pagerduty.com"
	}
	return "https://events.pagerduty.com"
}

func PdApiUrl() string {
	region := viper.GetString("region")
	if region == "eu" {
		return "https://api.eu.pagerduty.com"
	}
	return "https://api.pagerduty.com"
}
