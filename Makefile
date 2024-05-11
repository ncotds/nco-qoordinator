GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

.PHONY: test
#? test: Run the unit and integration tests
test: test-unit

.PHONY: test-unit
#? test-unit: Run the unit tests
test-unit:
	GOOS=$(GOOS) GOARCH=$(GOARCH) gotestsum --junitfile=coverage.xml -- -coverprofile=coverage.txt -covermode atomic -race ./pkg/... ./cmd/...

.PHONY: lint
#? lint: Run golangci-lint
lint:
	golangci-lint run ./...
