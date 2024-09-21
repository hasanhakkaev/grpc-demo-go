# ==================================================================================== #
# HELPERS
# ==================================================================================== #
DB_DSN ?= postgres:postgres@localhost:5432/postgres?sslmode=disable

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## Run producer
.PHONY: producer
producer:
	@echo "Running task producer.."
	@go run cmd/producer/*.go

## Run consumer
.PHONY: consumer
consumer:
	@echo "Running task consumer.."
	@go run cmd/consumer/main.go


## Test
.PHONY: test
test:
	@docker compose -f test/docker-compose.yml down -v
	@docker compose -f test/docker-compose.yml up --build --abort-on-container-exit --remove-orphans --force-recreate
	@docker compose -f test/docker-compose.yml down -v

## Stack
.PHONY:	stop
stop:
	@docker compose -f stack.yml down -v

.PHONY:	prod
prod:
	@docker compose -f stack.yml down -v
	@docker compose -f stack.yml up --build

.PHONY: dev
dev:
	@docker compose -f stack.yml down -v
	@docker compose -f stack.yml -f stack.dev.yml up

.PHONY: generate
generate:
	@sqlc generate
	@protoc --proto_path=proto proto/*.proto  --go_out=:. --go-grpc_out=:.


## db-up: start database
.PHONY: db-up
db-up:
	@ echo "Starting database ..."
	@docker run --rm -it --name yqapp-demo_db -e POSTGRES_PASSWORD=postgres -d -p 5432:5432 postgres:13
## db-down: stop database
db-down:
	@ echo "Stopping database ..."
	@docker stop yqapp-demo_db

## migrations/up: apply all up database migrations
.PHONY: migrations/up
migrations/up:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path=./assets/migrations -database="postgres://${DB_DSN}" up

## migrations/down: apply all down database migrations
.PHONY: migrations/down
migrations/down:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path=./assets/migrations -database="postgres://${DB_DSN}" down

## monitoring-up: start Monitoring stack
monitoring-up:
	@$HOSTNAME=$(hostname) docker stack deploy -c monitoring/docker-stack.yaml prom