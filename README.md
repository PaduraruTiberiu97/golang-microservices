# Go Microservices Project

This repository is a Go microservices playground with six services and two deployment modes (Docker Compose and Kubernetes).

## What The App Does

- `front-end`: serves a test UI to trigger workflows.
- `broker-service`: central API gateway; routes actions to downstream services.
- `authentication-service`: validates user credentials against PostgreSQL.
- `logger-service`: stores log events in MongoDB, exposed over HTTP, RPC, and gRPC.
- `mail-service`: sends email via SMTP (typically MailHog in local dev).
- `listener-service`: consumes RabbitMQ topic events and forwards them to logger.

## Service Communication

- Front-end -> Broker via HTTP.
- Broker -> Authentication via HTTP.
- Broker -> Logger via RPC (default) or gRPC (`/log-grpc` endpoint).
- Broker -> Mailer via HTTP.
- Authentication -> Logger via HTTP (login event logging).
- Listener -> RabbitMQ topic exchange (`logs_topic`) -> Logger via HTTP.

## Ports And Endpoints

- `front-end` (default local): `:8081`
  - `GET /`
- `broker-service` (container port `80`, mapped to host `8000` in Compose):
  - `POST /`
  - `POST /handle`
  - `POST /log-grpc`
  - `GET /ping`
- `authentication-service` (container port `80`, mapped to host `8081` in Compose):
  - `POST /authenticate`
  - `GET /ping`
- `logger-service`:
  - HTTP `:80`: `POST /log`, `GET /ping`
  - RPC `:5001`: `RPCServer.LogInfo`
  - gRPC `:50001`: `logs.LogService/Write`
- `mail-service` (container port `80`):
  - `POST /send`
  - `GET /ping`
- `listener-service`: no HTTP API; background consumer.

## Required Runtime Dependencies

- PostgreSQL (for `authentication-service`)
- MongoDB (for `logger-service`)
- RabbitMQ (for broker/listener event flow)
- MailHog or SMTP server (for `mail-service`)

## Environment Variables

### `front-end`
- `BROKER_URL` (example: `http://localhost:8000`)

### `authentication-service`
- `DSN` (PostgreSQL DSN)

### `mail-service`
- `MAIL_DOMAIN`
- `MAIL_HOST`
- `MAIL_PORT`
- `MAIL_ENCRYPTION` (`none`, `tls`, `ssl`)
- `MAIL_USERNAME`
- `MAIL_PASSWORD`
- `MAIL_NAME`
- `MAIL_ADDRESS`

### `logger-service`
- Mongo credentials are currently configured in code (`admin/password`) and expected to match container config.

## Run Locally (Docker Compose + Front-End Binary)

1. Build service binaries and start Compose stack:
```bash
cd project
make up_build
```

2. Run front-end locally (outside compose):
```bash
cd ../front-end
BROKER_URL=http://localhost:8000 go run .
```

3. Open:
- `http://localhost:8081` (front-end UI)
- `http://localhost:8025` (MailHog UI, if running through compose)

4. Stop stack:
```bash
cd ../project
make down
```

## Kubernetes Manifests

Kubernetes manifests are in `project/k8s/`, and ingress rules are in `project/ingress.yml`.

Typical flow:
```bash
kubectl apply -f project/k8s
kubectl apply -f project/ingress.yml
```

## Development Commands

- Format Go code:
```bash
find authentication-service broker-service front-end listener-service logger-service mail-service -type f -name '*.go' ! -name '*.pb.go' -print0 | xargs -0 gofmt -w
```

- Run tests (currently auth-service has tests):
```bash
cd authentication-service && go test ./...
```

## Documentation

- Project-wide file reference: `docs/FILES.md`

## Notes / Caveats In Current Code

- Broker mail routing uses `http://mailer-service/send`; in Docker Compose the service name is `mail-service` (Kubernetes manifest uses `mailer-service`).
- Repository currently contains built binaries (`authApp`, `brokerApp`, etc.) and runtime DB volume data under `project/db-data/`.
