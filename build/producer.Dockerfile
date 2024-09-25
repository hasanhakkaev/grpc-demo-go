FROM golang:1.23.1-alpine AS builder

# Install necessary build dependencies (use minimal packages to reduce build time)
RUN apk add --no-cache gcc musl-dev

# Set the build environment
WORKDIR /build

# Copy go.mod and go.sum first (this step benefits from caching)
COPY go.mod go.sum ./

# Download Go modules (this step will be cached as long as go.mod and go.sum don’t change)
RUN go mod download -x

# Now copy the entire source code (this invalidates the cache only if the source code changes)
COPY . .

# Build the Go binary (you can optimize with specific flags for performance or size)
ARG BUILD_TIME
ARG PRODUCER_VERSION
RUN CGO_ENABLED=0 go build -a -ldflags "-s -w -X main.Version=${PRODUCER_VERSION} -X main.BuildTime=${BUILD_TIME}" -o producer cmd/producer/main.go

# Stage 2: Minimal runtime image
FROM alpine:latest

# Copy the built Go binary from the builder stage
COPY --from=builder /build/producer /producer

# Set environment variables (you can pass them from the build)
ENV PRODUCER_VERSION=${PRODUCER_VERSION}
ENV BUILD_TIME=${BUILD_TIME}
ENV CONFIG_ENV=${CONFIG_ENV}

# Set the user to a non-root user for security
USER 65534

# Run the application
CMD ["/producer"]