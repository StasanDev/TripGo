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

.PHONY: build run test test-integration lint generate migrate migrate-down migrate-status

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
