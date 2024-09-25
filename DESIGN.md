# Task Application - Design document

# Introduction

This document contains the technical documentation for the Task application. This document provides an overview of the
architectural choices and design principles followed throughout the development of this project.

Technology stack:
## 1. gRPC for Inter-Service Communication:
- **Chosen over REST** because gRPC offers a more **efficient, binary protocol** with better support for streaming and handling high-throughput communication. Given the requirement for thousands of tasks per second, gRPC's performance and **built-in support for protobuf serialization** made it the optimal choice.
- **Reasoning**: gRPC's **low-latency and high-performance** nature make it well-suited for microservices where real-time task processing and high-throughput are required.
## 2. PostgreSQL for Task Persistence:
- **Reasoning**: PostgreSQL’s **ACID properties** and **transaction support** made it the most suitable choice for ensuring that tasks are not lost and are persisted even if the consumer fails. Its support for complex queries also allows for flexible task management (e.g., retrieving tasks by state).
## 3. uber-go/zap as Logging library:
- **Chosen over other alternatives due to**
- ** Structured Logging
- ** Performance
- ** Compatibility with Observability Tools
- ** Simple and Flexible Configuration
- ** Wide Adoption and Support
- 
## 4. go-migrate for Database migrations
- **Reasoning**
- Provides seemless and easy integration via library
- Has a cli tool if desired


## 5.Code generation
- [Protocol buffers](https://protobuf.dev/) as single source of truth. Proto files can be found under the `proto/` folder.
- [SQLC](https://github.com/sqlc-dev/sqlc) for type-safe code from SQL.
