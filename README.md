# Go Microservices

A Go microservices playground with 6 services, message-driven logging, and two deployment modes:
- Local stack with Docker Compose
- Kubernetes manifests + ingress

## Architecture

```text
Front-end (UI)
   |
   v
Broker Service (HTTP)
   |---> Authentication Service (HTTP, Postgres)
   |---> Logger Service (RPC by default, or gRPC)
   |---> Mail Service (HTTP, SMTP/Mailpit)
   |
   `---> RabbitMQ (optional event path)
             |
             v
        Listener Service ---> Logger Service (HTTP)
```

## Services

| Service | Purpose | Interface |
|---|---|---|
| `front-end` | UI to trigger broker workflows | HTTP `GET /` |
| `broker-service` | API gateway/orchestrator | HTTP `POST /`, `POST /handle`, `POST /log-grpc` |
| `authentication-service` | Credential validation | HTTP `POST /authenticate` |
| `logger-service` | Persist logs to MongoDB | HTTP `POST /log`, RPC `LogInfo`, gRPC `Write` |
| `mail-service` | SMTP email sender | HTTP `POST /send` |
| `listener-service` | RabbitMQ topic consumer forwarding to logger | background worker |

## Prerequisites

- Go (matching module versions used in the repo)
- Docker + Docker Compose
- `make`
- Optional for Kubernetes:
  - `kubectl`
  - local cluster (for example Minikube)
  - ingress controller enabled

## Quick Start (Docker Compose + local UI)

This is the easiest way to run everything.

1. Start backend stack:

```bash
cd project
make up_build
```

2. Start front-end (runs on `8082` by default to avoid collisions):

```bash
cd project
make start
```

3. Open:

- Front-end UI: `http://localhost:8082`
- Broker API: `http://localhost:8000`
- Mail inbox (Mailpit): `http://localhost:8025`
- Auth service (direct): `http://localhost:8081`

4. Stop front-end:

```bash
cd project
make stop
```

5. Stop backend stack:

```bash
cd project
make down
```

## Front-end Runtime Options

The front-end supports these environment variables:

- `FRONTEND_PORT` (default: `8081`)
- `BROKER_URL` (default: `http://localhost:8000`)
- `MAIL_INBOX_URL` (default: `http://localhost:8025`)

Example (manual run without Makefile):

```bash
cd front-end
FRONTEND_PORT=8082 BROKER_URL=http://localhost:8000 MAIL_INBOX_URL=http://localhost:8025 go run .
```

## Broker API Examples

Health-style broker call:

```bash
curl -s -X POST http://localhost:8000 | jq
```

Auth flow via broker:

```bash
curl -s -X POST http://localhost:8000/handle \
  -H 'Content-Type: application/json' \
  -d '{"action":"auth","auth":{"email":"admin@example.com","password":"verysecret"}}' | jq
```

Log via RPC path (default `/handle` log action):

```bash
curl -s -X POST http://localhost:8000/handle \
  -H 'Content-Type: application/json' \
  -d '{"action":"log","log":{"name":"event","data":"hello from curl"}}' | jq
```

Log via gRPC path (`/log-grpc` endpoint):

```bash
curl -s -X POST http://localhost:8000/log-grpc \
  -H 'Content-Type: application/json' \
  -d '{"action":"log","log":{"name":"event","data":"hello via grpc endpoint"}}' | jq
```

Mail flow via broker:

```bash
curl -s -X POST http://localhost:8000/handle \
  -H 'Content-Type: application/json' \
  -d '{"action":"mail","mail":{"from":"sender@example.com","to":"recipient@example.com","subject":"Test","message":"Hello"}}' | jq
```

## Environment Variables by Service

### `broker-service`

- `AUTH_SERVICE_URL` (default: `http://authentication-service/authenticate`)
- `MAIL_SERVICE_URL` (default: `http://mail-service/send`)
- `LOGGER_SERVICE_URL` (default: `http://logger-service/log`)
- `LOGGER_RPC_ADDR` (default: `logger-service:5001`)
- `LOGGER_GRPC_ADDR` (default: `logger-service:50001`)
- `RABBITMQ_URL` (default: `amqp://guest:guest@rabbitmq`)

### `authentication-service`

- `DSN` (default in code: `host=postgres port=5432 user=postgres password=password dbname=users sslmode=disable timezone=UTC connect_timeout=5`)
- `LOGGER_SERVICE_URL` (default: `http://logger-service/log`)

### `logger-service`

- `MONGO_URI` (default: `mongodb://mongo:27017`)
- `MONGO_INITDB_ROOT_USERNAME` (default: `admin`)
- `MONGO_INITDB_ROOT_PASSWORD` (default: `password`)

### `mail-service`

- `MAIL_DOMAIN`
- `MAIL_HOST`
- `MAIL_PORT`
- `MAIL_ENCRYPTION` (`none`, `tls`, `ssl`)
- `MAIL_USERNAME`
- `MAIL_PASSWORD`
- `MAIL_NAME`
- `MAIL_ADDRESS`

### `listener-service`

- `LOGGER_SERVICE_URL` (default: `http://logger-service/log`)
- `RABBITMQ_URL` (default: `amqp://guest:guest@rabbitmq`)

## Kubernetes

Manifests are under `project/k8s/` and ingress is in `project/ingress.yml`.

Apply:

```bash
kubectl apply -f project/k8s
kubectl apply -f project/ingress.yml
```

Ingress hosts:

- `http://front-end.127.0.0.1.nip.io`
- `http://broker-service.127.0.0.1.nip.io`

Delete everything:

```bash
kubectl delete -f project/ingress.yml --ignore-not-found
kubectl delete -f project/k8s --ignore-not-found
```

## Development Commands

Format Go files:

```bash
find authentication-service broker-service front-end listener-service logger-service mail-service \
  -type f -name '*.go' ! -name '*.pb.go' -print0 | xargs -0 gofmt -w
```

Run tests across modules:

```bash
cd authentication-service && go test ./...
cd ../broker-service && go test ./...
cd ../logger-service && go test ./...
cd ../mail-service && go test ./...
cd ../listener-service && go test ./...
cd ../front-end && go test ./...
```

Run vet across modules:

```bash
cd authentication-service && go vet ./...
cd ../broker-service && go vet ./...
cd ../logger-service && go vet ./...
cd ../mail-service && go vet ./...
cd ../listener-service && go vet ./...
cd ../front-end && go vet ./...
```

## Troubleshooting

- `listen tcp :8081: bind: address already in use` when starting front-end:
  - `authentication-service` uses host `8081` in Compose.
  - Use `make start` (defaults to `8082`) or set `FRONTEND_PORT` explicitly.

- Front-end cannot reach broker:
  - Verify `BROKER_URL` (default `http://localhost:8000`).
  - Verify backend stack is running: `cd project && docker compose ps`.

- Mail inbox empty:
  - Verify Mailpit container is running and front-end inbox URL is correct (`MAIL_INBOX_URL`).

## Additional Documentation

- Full file-by-file technical inventory: `docs/FILES.md`
