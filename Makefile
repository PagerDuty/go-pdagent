build: test pagerduty-agent

GIT_COMMIT = $(shell git rev-list -1 HEAD)

pagerduty-agent:
	go build -ldflags "-s -w -X common.GitCommit=$(GIT_COMMIT)" .

.PHONY: test
test:
	go test ./...

.PHONY: format
format:
	go fmt ./...

clean:
	rm -f pagerduty-agent
