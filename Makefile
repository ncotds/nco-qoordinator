GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

API_CONFIG_PATH=$(shell pwd)/local.yml

.PHONY: setup-tools
#? setup-tools: Install dev tools
setup-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.58.0
	go install github.com/michurin/human-readable-json-logging/cmd/pplog@v0.0.0-20240616030539-dd4a67e261f0
	go install github.com/vektra/mockery/v2@v2.43.0
	go install github.com/ogen-go/ogen/cmd/ogen@v1.2.2
	go install golang.org/x/tools/cmd/goimports@v0.21.0
	go install gotest.tools/gotestsum@v1.11.0

.PHONY: test
#? test: Run the unit and integration tests
test: test-unit

.PHONY: test-unit
#? test-unit: Run the unit tests
test-unit:
	GOOS=$(GOOS) GOARCH=$(GOARCH) gotestsum --junitfile=coverage-unit.xml --jsonfile=coverage-unit.json -- \
 		-coverprofile=coverage-unit.txt -covermode atomic -race  ./pkg/... ./cmd/... ./internal/...

.PHONY: test-e2e
#? test-e2e: Run the E2E tests
test-e2e:
	GOOS=$(GOOS) GOARCH=$(GOARCH) gotestsum -- -race ./tests/e2e/...

.PHONY: fmt
#? fmt: Run gofmt
fmt:
	gofmt -s -l -w internal/ pkg/ cmd/ tests/e2e/

.PHONY: lint
#? lint: Run golangci-lint
lint:
	golangci-lint run ./...

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
	pplog go run ./cmd/ncoq-api/. -c $(API_CONFIG_PATH)