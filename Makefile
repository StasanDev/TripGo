-include .env

HTTP_READ_TIMEOUT ?= 30s
HTTP_READ_HEADER_TIMEOUT ?= 5s
HTTP_WRITE_TIMEOUT ?= 30s
HTTP_IDLE_TIMEOUT ?= 60s
IDEMPOTENCY_TTL ?= 24h
OTEL_METRIC_EXPORT_INTERVAL ?= 60000
OTEL_METRIC_EXPORT_TIMEOUT ?= 30000

export HTTP_ADDR
export HTTP_READ_TIMEOUT
export HTTP_READ_HEADER_TIMEOUT
export HTTP_WRITE_TIMEOUT
export HTTP_IDLE_TIMEOUT
export LOG_LEVEL
export SHUTDOWN_TIMEOUT
export IDEMPOTENCY_TTL
export DATABASE_URL
export DATABASE_MAX_CONNS
export DATABASE_MIN_CONNS
export DATABASE_MAX_CONN_LIFETIME
export DATABASE_CONNECT_TIMEOUT
export DATABASE_QUERY_TIMEOUT
export OTEL_SERVICE_NAME
export OTEL_EXPORTER_OTLP_ENDPOINT
export OTEL_EXPORTER_OTLP_PROTOCOL
export OTEL_RESOURCE_ATTRIBUTES
export OTEL_TRACES_SAMPLER
export OTEL_METRIC_EXPORT_INTERVAL
export OTEL_METRIC_EXPORT_TIMEOUT

GOOSE_DRIVER = postgres
MIGRATIONS_DIR = migrations
OPENAPI_SPEC = contracts/openapi/trip-service.openapi.yaml
OPENAPI_OUTPUT = internal/generated/api.gen.go
OPENAPI_OPERATIONS = createTrip,getTrip,finishTrip,health,ready
COVERAGE_FILE = coverage.out
COVERAGE_PACKAGES = ./internal/domain/... ./internal/usecase/...
INTEGRATION_PACKAGES = ./internal/repository/postgres/...
SERVICE_VERSION ?= dev
GO_LDFLAGS = -X main.serviceVersion=$(SERVICE_VERSION)

MUTATION_STAGE ?= current
MUTATION_DIR := $(CURDIR)/docs/lab2/mutation

GREMLINS := $(CURDIR)/bin/gremlins
GREMLINS_FLAGS := --invert-logical --invert-assignments \
	--invert-bitwise --invert-bwassign --invert-loopctrl \
	--remove-self-assignments --workers 2 --timeout-coefficient 10

GO_MUTESTING := $(CURDIR)/bin/go-mutesting
GO_MUTESTING_VERSION := v0.0.0-20210610104036-6d9217011a00
GO_MUTESTING_FLAGS := --exec-timeout=60

.PHONY: build run test test-integration lint generate migrate migrate-down migrate-status mutation

build:
	go build -ldflags "$(GO_LDFLAGS)" ./...

run:
	go run -ldflags "$(GO_LDFLAGS)" ./cmd/trip-service

test:
	go test -race -count=2 ./...
	go test -race -covermode=atomic -coverprofile=$(COVERAGE_FILE) $(COVERAGE_PACKAGES)
	go tool cover -func=$(COVERAGE_FILE)

test-integration:
	go test -race -count=2 -tags=integration $(INTEGRATION_PACKAGES)

lint:
	golangci-lint run --config .golangci.yml

generate:
	@mkdir -p "$(dir $(OPENAPI_OUTPUT))"
	go tool oapi-codegen \
		-generate types,chi-server,spec \
		-include-operation-ids $(OPENAPI_OPERATIONS) \
		-package api \
		-o "$(OPENAPI_OUTPUT)" \
		"$(OPENAPI_SPEC)"
	go generate ./...

migrate:
	go tool goose -dir "$(MIGRATIONS_DIR)" "$(GOOSE_DRIVER)" "$(DATABASE_URL)" up

migrate-down:
	go tool goose -dir "$(MIGRATIONS_DIR)" "$(GOOSE_DRIVER)" "$(DATABASE_URL)" down

migrate-status:
	go tool goose -dir "$(MIGRATIONS_DIR)" "$(GOOSE_DRIVER)" "$(DATABASE_URL)" status

$(GREMLINS):
	GOBIN="$(CURDIR)/bin" go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0

# The original x/tools dependency is incompatible with Go 1.27.
# Build the pinned tool with a compatible x/tools in an isolated module.
$(GO_MUTESTING):
	@set -eu; \
		build_dir=$$(mktemp -d); \
		trap 'rm -rf "$$build_dir"' EXIT; \
		cd "$$build_dir"; \
		go mod init mutation-tools; \
		go get github.com/zimmski/go-mutesting/cmd/go-mutesting@$(GO_MUTESTING_VERSION) golang.org/x/tools@v0.51.0; \
		GOBIN="$(@D)" go install github.com/zimmski/go-mutesting/cmd/go-mutesting

mutation: SHELL := bash
mutation: .SHELLFLAGS := -o pipefail -c
mutation: $(GREMLINS) $(GO_MUTESTING)
	@command -v jq >/dev/null
	mkdir -p "$(MUTATION_DIR)"
	"$(GREMLINS)" unleash "$(CURDIR)/internal/domain" $(GREMLINS_FLAGS) \
		--output "$(MUTATION_DIR)/domain-$(MUTATION_STAGE).json"
	jq --indent 4 . "$(MUTATION_DIR)/domain-$(MUTATION_STAGE).json" > "$(MUTATION_DIR)/domain-$(MUTATION_STAGE).json.tmp"
	mv "$(MUTATION_DIR)/domain-$(MUTATION_STAGE).json.tmp" "$(MUTATION_DIR)/domain-$(MUTATION_STAGE).json"
	"$(GREMLINS)" unleash "$(CURDIR)/internal/usecase" $(GREMLINS_FLAGS) \
		--output "$(MUTATION_DIR)/usecase-$(MUTATION_STAGE).json"
	jq --indent 4 . "$(MUTATION_DIR)/usecase-$(MUTATION_STAGE).json" > "$(MUTATION_DIR)/usecase-$(MUTATION_STAGE).json.tmp"
	mv "$(MUTATION_DIR)/usecase-$(MUTATION_STAGE).json.tmp" "$(MUTATION_DIR)/usecase-$(MUTATION_STAGE).json"
	"$(GO_MUTESTING)" $(GO_MUTESTING_FLAGS) ./internal/domain/... ./internal/usecase/... \
		2>&1 | tee "$(MUTATION_DIR)/go-mutesting-$(MUTATION_STAGE).log"
