# ==================================================================================== #
# HELPERS
# ==================================================================================== #
DB_DSN ?= postgres:postgres@localhost:5432/postgres?sslmode=disable

PRODUCER_VERSION := $(shell git rev-parse --short HEAD)
CONSUMER_VERSION := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)

PPROF_HOST ?= localhost
PPROF_PRODUCER_PORT ?= 6061
PPROF_CONSUMER_PORT ?= 6060
FLAMEGRAPH_DIR = /path/to/FlameGraph  # Update this path to where you cloned FlameGraph
PROFILE_DURATION ?= 30s  # Default profile duration

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
	@GOGC=50 GOMEMLIMIT=4096MiB go run cmd/producer/main.go

## Run consumer
.PHONY: run/consumer
run/consumer:
	@echo "Running task consumer.."
	@GOGC=50 GOMEMLIMIT=4096MiB go run cmd/consumer/main.go

## Build producer
.PHONY: build/producer
build/producer:
	@echo "Building task producer.."
	@go build -ldflags="-s -w -X main.Version=$(PRODUCER_VERSION) -X main.BuildTime=$(BUILD_TIME)" -o producer cmd/producer/main.go
	@ls -lah producer

## Build consumer
.PHONY: build/consumer
build/consumer:
	@echo "Building task consumer.."
	@go build  -ldflags="-s -w -X main.Version=$(CONSUMER_VERSION) -X main.BuildTime=$(BUILD_TIME)"  -o consumer cmd/consumer/main.go
	@ls -lah consumer

### Deploy docker
.PHONY: deploy
deploy:
	@echo "Deploying locally with docker-compose"
	@DOCKER_BUILDKIT=1 \
	PRODUCER_VERSION=$(PRODUCER_VERSION)\
 	CONSUMER_VERSION=$(CONSUMER_VERSION)\
 	BUILD_TIME=${BUILD_TIME}\
 	docker-compose -f deployment/docker/docker-compose.yaml build

	@docker-compose -f deployment/docker/docker-compose.yaml up --abort-on-container-exit --remove-orphans --force-recreate

.PHONY: view-flamegraph-consumer
view-flamegraph-consumer:
	@echo "Opening flamegraph in the browser..."
	@open cpu_flamegraph.svg

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
start/services:
	@docker-compose -f deployment/docker/services.yaml up -d --remove-orphans

## stop/infra: stop Docker Stack
stop/services:
	@docker-compose  -f deployment/docker/services.yaml down