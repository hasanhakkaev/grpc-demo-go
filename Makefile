# ==================================================================================== #
# HELPERS
# ==================================================================================== #
DB_DSN ?= postgres:postgres@localhost:5432/postgres?sslmode=disable

PRODUCER_VERSION := $(shell git rev-parse --short HEAD)
CONSUMER_VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date +%Y-%m-%dT%H:%M:%S)

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

lint:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --print-issued-lines=true

## Run producer
.PHONY: run/producer
run/producer:
	@echo "Running task producer.."
	@go run cmd/producer/main.go

## Run consumer
.PHONY: run/consumer
run/consumer:
	@echo "Running task consumer.."
	@go run cmd/consumer/main.go

## Build producer
.PHONY: build/producer
build/producer:
	@echo "Building task producer.."
	@go build -ldflags="-s -w -X main.Version=$(PRODUCER_VERSION) -X main.BuildTime=$(BUILDTIME)" -o producer cmd/producer/main.go
	@ls -lah producer

## Build consumer
.PHONY: build/consumer
build/consumer:
	@echo "Building task consumer.."
	@go build  -ldflags="-s -w -X main.Version=$(CONSUMER_VERSION) -X main.BuildTime=$(BUILDTIME)"  -o consumer cmd/consumer/main.go
	@ls -lah consumer


### Test
#.PHONY: test
#test:
#	@docker compose -f test/docker-compose.yml down -v
#	@docker compose -f test/docker-compose.yml up --build --abort-on-container-exit --remove-orphans --force-recreate
#	@docker compose -f test/docker-compose.yml down -v
#
### Stack
#.PHONY:	stop
#stop:
#	@docker compose -f stack.yml down -v
#
#.PHONY:	prod
#prod:
#	@docker compose -f stack.yml down -v
#	@docker compose -f stack.yml up --build
#
#.PHONY: dev
#dev:
#	@docker compose -f stack.yml down -v
#	@docker compose -f stack.yml -f stack.dev.yml up

.PHONY: generate
generate:
	@sqlc generate
	@protoc --proto_path=proto proto/*.proto  --go_out=:. --go-grpc_out=:.

## migrations/up: apply all up database migrations
.PHONY: migrations/up
migrations/up:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path=./assets/migrations -database="postgres://${DB_DSN}" up

## migrations/down: apply all down database migrations
.PHONY: migrations/down
migrations/down:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path=./assets/migrations -database="postgres://${DB_DSN}" down

## start/infra: start Docker stack
start/infra:
	@docker-compose -f deployment/local/docker-compose.yaml up -d

## stop/infra: stop Docker Stack
stop/infra:
	@docker-compose  -f deployment/local/docker-compose.yaml down