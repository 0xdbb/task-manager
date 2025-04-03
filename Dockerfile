# Use the official Go image to build the application
FROM golang:1.23.8-alpine3.20 AS builder

WORKDIR /app
# Install goose (Database Migration Tool)
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

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
COPY .env ./

# Copy goose binary from the builder stage
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY .env ./
COPY Makefile ./
COPY entrypoint.sh /app/entrypoint.sh
COPY internal/database/migrations/ /app/internal/database/migrations
RUN chmod +x /app/entrypoint.sh

# Expose the port for the API
EXPOSE 8000

ENTRYPOINT []

CMD ["/app/main"]
