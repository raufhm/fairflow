# Multi-stage build for Analytics Service
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY shared/ ./shared/
COPY services/analytics/ ./services/analytics/
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/analytics-service ./services/analytics/cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/analytics-service .
EXPOSE 3007
CMD ["./analytics-service"]
