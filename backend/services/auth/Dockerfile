# Multi-stage build for Auth Service
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy shared packages and service code
COPY shared/ ./shared/
COPY services/auth/ ./services/auth/

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/auth-service ./services/auth/cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/auth-service .

# Expose service port
EXPOSE 3001

# Run the service
CMD ["./auth-service"]
