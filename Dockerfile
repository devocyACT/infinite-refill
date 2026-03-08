# Multi-stage build
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
ARG VERSION=dev
ARG BUILD_TIME
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o /app/refill ./cmd/refill

# Final image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary
COPY --from=builder /app/refill /usr/local/bin/refill

# Create working directory
WORKDIR /data

# Set timezone
ENV TZ=UTC

# Health check
HEALTHCHECK --interval=5m --timeout=3s \
  CMD refill check || exit 1

# Default command
ENTRYPOINT ["refill"]
CMD ["scheduler", "start"]
