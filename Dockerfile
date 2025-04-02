# Use the official Go image to build the application
FROM golang:1.23.8-alpine3.20 AS builder

WORKDIR /app

# Copy go mod files first and download dependencies (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build both API and Worker in one step
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/api/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/worker ./cmd/worker/main.go

# Use a minimal base image
FROM alpine:latest

WORKDIR /app

# Install necessary system dependencies
RUN apk --no-cache add ca-certificates

# Copy the built binaries
COPY --from=builder /app/main /app/main
COPY --from=builder /app/worker /app/worker

# Expose the port for the API
EXPOSE 8080

# Allow the container to run API or worker dynamically
ENTRYPOINT []
CMD ["/app/main"]

