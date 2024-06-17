GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

GOTESTSUM_CMD=go run gotest.tools/gotestsum
GOLANGCI_LINT_CMD=go run github.com/golangci/golangci-lint/cmd/golangci-lint

API_CONFIG_PATH=$(shell pwd)/local.yml

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

.PHONY: test-e2e
#? test-e2e: Run the E2E tests
test-e2e:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOTESTSUM_CMD) --junitfile=coverage.xml -- -tags=integration -coverprofile=coverage.txt -covermode atomic -race ./tests/e2e/...

.PHONY: lint
#? lint: Run golangci-lint
lint:
	gofmt -s -l -w internal/ pkg/ cmd/ tests/e2e/
	$(GOLANGCI_LINT_CMD) run ./...

.PHONY: generate
#? generate: Run go generate
generate:
	go generate ./...

.PHONY: benchmark-restapi
#? benchmark-restapi: Run restapi benchmarks
benchmark-restapi:
	go test -bench . -benchmem ./internal/restapi/.

.PHONY: benchmark-e2e
#? benchmark-e2e: Run e2e benchmarks
benchmark-e2e:
	go test -bench . -benchmem ./tests/e2e/...

.PHONY: run-ncoq-api
#? run-ncoq-api: Run cmd/ncoq-api
run-ncoq-api:
	go run github.com/vitalyshatskikh/human-readable-json-logging/cmd/pplog go run ./cmd/ncoq-api/. -c $(API_CONFIG_PATH)