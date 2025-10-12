# Multi-stage build for Member Service
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy shared packages and service code
COPY shared/ ./shared/
COPY services/member/ ./services/member/

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/member-service ./services/member/cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/member-service .

# Expose service port
EXPOSE 3003

# Run the service
CMD ["./member-service"]
