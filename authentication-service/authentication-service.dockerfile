# Runtime image for authentication-service.
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/authApp ./cmd/api

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/authApp /app/authApp
ENTRYPOINT ["/app/authApp"]
