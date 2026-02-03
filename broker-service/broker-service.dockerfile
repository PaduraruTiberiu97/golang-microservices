FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
RUN mkdir /app

# Copy the Go module files and download dependencies
COPY . /app

# Set the working directory to /app
WORKDIR /app

# Build the Go application
RUN CGO_ENABLED=0 go build -o brokerApp ./cmd/api

# Use a minimal base image for the final container
RUN chmod +x /app/brokerApp

# Start a new stage from a minimal base image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/brokerApp /app

CMD ["/app/brokerApp"]
