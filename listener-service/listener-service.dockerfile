# Runtime image for listener-service.
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/listenerApp .

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/listenerApp /app/listenerApp
ENTRYPOINT ["/app/listenerApp"]
