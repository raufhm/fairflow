# Multi-stage build for Group Service
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy shared packages and service code
COPY shared/ ./shared/
COPY services/group/ ./services/group/

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/group-service ./services/group/cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/group-service .

# Expose service port
EXPOSE 3002

# Run the service
CMD ["./group-service"]
