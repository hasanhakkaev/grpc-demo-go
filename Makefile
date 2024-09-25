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
FLAMEGRAPH_DIR = /tmp/FlameGraph  # Update this path to where you cloned FlameGraph
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
	@go run cmd/producer/main.go

## Run consumer
.PHONY: run/consumer
run/consumer:
	@echo "Running task consumer.."
	@go run cmd/consumer/main.go

## Run producer
.PHONY: run/producer/512
run/producer/512:
	@echo "Running task producer.."
	@GOGC=25 GOMEMLIMIT=2048MiB go run cmd/producer/main.go

## Run consumer
.PHONY: run/consumer/512
run/consumer/512:
	@echo "Running task consumer.."
	@GOGC=25 GOMEMLIMIT=2048MiB go run cmd/consumer/main.go


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

## Docker build
.PHONY: docker/build
docker/build:
	@echo "Build producer and consumer containers"
	@DOCKER_BUILDKIT=1 \
	PRODUCER_VERSION=$(PRODUCER_VERSION)\
 	CONSUMER_VERSION=$(CONSUMER_VERSION)\
 	BUILD_TIME=${BUILD_TIME}\
 	docker-compose -f deployment/docker/docker-compose.yaml build

## Docker run
docker/run:
	@echo "Running entire stack with db and monitoring"
	@@docker-compose -f deployment/docker/docker-compose.yaml up -d  --remove-orphans --force-recreate

.PHONY: view-flamegraph-consumer
view-flamegraph-consumer:
	@echo "Opening flamegraph in the browser..."
	@open cpu_flamegraph.svg

### FlameGraph Install
#.PHONY: flamegraph/install
#flamegraph/install:
#	@git clone https://github.com/brendangregg/FlameGraph.git ${FLAMEGRAPH_DIR} || echo "FlameGraph already cloned"
#
#.PHONY: flamegraph/consumer
#flamegraph/consumer:
#	@echo "Capturing CPU profile for $(PROFILE_DURATION)..."
#	@go tool pprof -raw  http://$(PPROF_HOST):$(PPROF_CONSUMER_PORT)/debug/pprof/profile?seconds=$(PROFILE_DURATION)  > cpu_profile.pb.gz
#	@echo "Converting pprof data to folded format..."
#	@go tool pprof -raw -output=cpu_profile.folded http://$(PPROF_HOST):$(PPROF_CONSUMER_PORT)/debug/pprof/profile
#	@echo "Generating flamegraph from folded data..."
#	@cat cpu_profile.folded | $(FLAMEGRAPH_DIR)/flamegraph.pl > cpu_flamegraph.svg
#	@echo "Flamegraph generated: cpu_flamegraph.svg"
#
#.PHONY: flamegraph/consumer-show
#flamegraph/consumer-show: flamegraph/consumer
#	@echo "Opening flamegraph in the browser..."
#	@open cpu_flamegraph.svg  # Or use xdg-open for Linux

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