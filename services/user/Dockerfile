# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy shared module
COPY services/shared /shared

# Copy user service source code
COPY services/user /app/services/user

# Fix the replace directive
WORKDIR /app/services/user
RUN go mod edit -replace=github.com/shevaathalla/posdigi-microservice/shared=/shared

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/user main.go app.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/services/user/bin/user .

# Expose port
EXPOSE 8002

# Run the application
CMD ["./user"]
