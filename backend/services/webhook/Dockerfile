# Multi-stage build for Webhook Service
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY shared/ ./shared/
COPY services/webhook/ ./services/webhook/
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/webhook-service ./services/webhook/cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/webhook-service .
EXPOSE 3005
CMD ["./webhook-service"]
