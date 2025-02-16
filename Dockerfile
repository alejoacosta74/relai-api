# Stage 1: Build the application
FROM golang:1.23.4-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/relai-api ./main.go

# Stage 2: Create the final image
FROM alpine:3.19

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/relai-api .

# Copy config file
COPY --from=builder /app/internal/config/config.yaml ./internal/config/config.yaml

# Create a non-root user
RUN adduser -D appuser
USER appuser

# Expose the port defined in your config
EXPOSE 8080

# Run the application
CMD ["./relai-api"]