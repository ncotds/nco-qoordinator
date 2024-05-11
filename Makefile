GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

GOTESTSUM_CMD=go run gotest.tools/gotestsum
GOLANGCI_LINT_CMD=go run github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: test
#? test: Run the unit and integration tests
test: test-unit

.PHONY: test-unit
#? test-unit: Run the unit tests
test-unit:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOTESTSUM_CMD) --junitfile=coverage.xml -- -coverprofile=coverage.txt -covermode atomic -race ./pkg/... ./cmd/...

.PHONY: lint
#? lint: Run golangci-lint
lint:
	$(GOLANGCI_LINT_CMD) run ./...
