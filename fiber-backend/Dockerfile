# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o plc-core ./cmd/server

# Production stage
FROM alpine:3.19 AS production

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

# Copy binary
COPY --from=builder /app/plc-core .

# Create directories
RUN mkdir -p /var/log/plc-core /exports /config

# Expose port
EXPOSE 3000

# Run
CMD ["./plc-core"]
