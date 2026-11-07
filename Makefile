-include .env

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

GOOSE_DRIVER = postgres
MIGRATIONS_DIR = migrations
OPENAPI_SPEC = contracts/openapi/trip-service.openapi.yaml
OPENAPI_OUTPUT = internal/generated/api.gen.go
OPENAPI_OPERATIONS = createTrip,getTrip,finishTrip,health,ready
COVERAGE_FILE = coverage.out
COVERAGE_PACKAGES = ./internal/domain/... ./internal/usecase/...
INTEGRATION_PACKAGES = ./internal/repository/postgres/...

.PHONY: build run test test-integration lint generate migrate migrate-down migrate-status

build:
	go build ./...

run:
	go run ./cmd/trip-service

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
