# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy shared module
COPY services/shared /shared

# Copy attendance service source code
COPY services/attendance /app/services/attendance

# Fix the replace directive
WORKDIR /app/services/attendance
RUN go mod edit -replace=github.com/shevaathalla/posdigi-microservice/shared=/shared

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/services/attendance/main .

# Expose port
EXPOSE 8003

# Run the application
CMD ["./main"]
