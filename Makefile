build: pagerduty-agent

GIT_COMMIT = $(shell git rev-list -1 HEAD)

pagerduty-agent: format test
	go build -ldflags "-s -w -X common.Commit=$(GIT_COMMIT)" .

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
	rm -f pagerduty-agent
