# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o noplacelike .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh noplacelike

# Set working directory
WORKDIR /home/noplacelike

# Copy binary from builder
COPY --from=builder /app/noplacelike .

# Create necessary directories
RUN mkdir -p uploads downloads plugins data && \
    chown -R noplacelike:noplacelike /home/noplacelike

# Switch to non-root user
USER noplacelike

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set environment variables
ENV NOPLACELIKE_HOST=0.0.0.0
ENV NOPLACELIKE_PORT=8080
ENV NOPLACELIKE_UPLOAD_FOLDER=/home/noplacelike/uploads
ENV NOPLACELIKE_DOWNLOAD_FOLDER=/home/noplacelike/downloads

# Run the application
CMD ["./noplacelike"]