build: pdagent

GIT_COMMIT = $(shell git rev-list -1 HEAD)
BUILD_DATE = $(shell date +%Y-%m-%dT%T%z)
BUILD_VERSION = $(shell git describe)

pdagent: test
	go build -o pdagent -ldflags "-s -w \
		-X 'github.com/PagerDuty/pagerduty-agent/pkg/common.Commit=$(GIT_COMMIT)' \
		-X 'github.com/PagerDuty/pagerduty-agent/pkg/common.Date=$(BUILD_DATE)' \
		-X 'github.com/PagerDuty/pagerduty-agent/pkg/common.Version=$(BUILD_VERSION)'" .

.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	go test ./...

.PHONY: release
release: format test
	goreleaser

.PHONY: release-test
release-test: format test
	goreleaser --snapshot --skip-publish --rm-dist

clean:
	rm -rf dist
	rm -f pdagent
	rm -f pagerduty-agent
