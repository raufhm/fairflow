# Multi-stage build for Assignment Service
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy shared packages and service code
COPY shared/ ./shared/
COPY services/assignment/ ./services/assignment/

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/assignment-service ./services/assignment/cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/assignment-service .

# Expose service port
EXPOSE 3004

# Run the service
CMD ["./assignment-service"]
