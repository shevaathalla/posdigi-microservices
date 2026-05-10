# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy gateway service go.mod and go.sum
COPY services/gateway/go.mod services/gateway/go.sum ./services/gateway/

WORKDIR /app/services/gateway
RUN go mod download

# Copy gateway service source code
COPY services/gateway .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/gateway main.go app.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/services/gateway/bin/gateway .

# Expose port
EXPOSE 8000

# Run the application
CMD ["./gateway"]
