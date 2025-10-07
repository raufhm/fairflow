# ============================================
# Build stage
# ============================================
FROM golang:1.25-alpine AS builder

# Install build dependencies (minimal - no CGO)
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

WORKDIR /app

# Copy go mod files first (better layer caching)
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build binaries with optimization flags
# CGO_ENABLED=0 for static binaries (no C dependencies)
# -ldflags for stripping debug info and reducing binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -trimpath \
    -o server \
    ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o migrate \
    ./cmd/migrate

# ============================================
# Runtime stage
# ============================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    && addgroup -g 1000 fairflow \
    && adduser -D -u 1000 -G fairflow fairflow

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/server .
COPY --from=builder /app/migrate .

# Copy migrations (required for migrate service)
# NOTE: These are embedded in the image for deployment only
# Source migrations remain private in repository
COPY --from=builder /app/migrations ./migrations

# Change ownership to non-root user
RUN chown -R fairflow:fairflow /app

# Switch to non-root user
USER fairflow

# Expose port
EXPOSE 3000

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Set environment variable
ENV PORT=3000

# Run the application
CMD ["./server"]