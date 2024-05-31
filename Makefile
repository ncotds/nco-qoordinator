GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

GOTESTSUM_CMD=go run gotest.tools/gotestsum
GOLANGCI_LINT_CMD=go run github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: test
#? test: Run the unit and integration tests
test: test-unit test-int

.PHONY: test-unit
#? test-unit: Run the unit tests
test-unit:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOTESTSUM_CMD) --junitfile=coverage.xml -- -coverprofile=coverage.txt -covermode atomic -race ./internal/... ./pkg/... ./cmd/...

.PHONY: test-int
#? test-unit: Run the integration tests
test-int:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOTESTSUM_CMD) --junitfile=coverage.xml -- -tags=integration -coverprofile=coverage.txt -covermode atomic -race ./internal/... ./pkg/... ./cmd/...

.PHONY: lint
#? lint: Run golangci-lint
lint:
	gofmt -s -l -w internal/ pkg/ cmd/
	$(GOLANGCI_LINT_CMD) run ./...

.PHONY: generate
#? generate: Run go generate
generate:
	go generate ./...

.PHONY: benchmark-restapi
#? benchmark-restapi: Run go generate
benchmark-restapi:
	go test -bench . -benchmem ./internal/restapi/.