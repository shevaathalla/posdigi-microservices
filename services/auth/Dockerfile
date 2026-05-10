# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy shared module
COPY services/shared /shared

# Copy auth service source code
COPY services/auth /app/services/auth

# Fix the replace directive
WORKDIR /app/services/auth
RUN go mod edit -replace=github.com/shevaathalla/posdigi-microservice/shared=/shared

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/auth main.go app.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/services/auth/bin/auth .

# Expose port
EXPOSE 8001

# Run the application
CMD ["./auth"]
