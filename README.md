# Task Application 

# Getting started

## Starting database and monitoring stack
```make start/services```
* Grafane is accessible on http://localhost:3000
* Prometheus s accessible on http://localhost:9090

## Stopping database and monitoring stack
```make stop/services```

## Code generation
```make generate```

## Database migration 
```make migrations/up```

## Database downgrade 
```task migrations/down```

## Running consumer
```make run/consumer```
* Metrcis endpoint http://localhost:4040/metrics
* Debug pprof endpoint http://localhost:6060/debug/pprof
* Consumer service uses port `50051` 

## Running producer
```make run/producer```
* Metrcis endpoint http://localhost:4041/metrics
* Debug pprof endpoint http://localhost:6061/debug/pprof

## Deploying as docker compose 
```make deploy```
* Builds both the producer and consumer
* Starts the whole stack together with db and monitoring

## Configuration
Local
- Can be found in `./configuration/` directory
- 

Docker 
- Can be found in `./deployment/docker/configuration`

