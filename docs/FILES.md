# File-By-File Documentation

This document covers the full source/configuration surface of the repository and explicitly classifies generated/runtime artifacts.

## Scope And Classification

- Source and config files: documented individually below.
- Generated code (`*.pb.go`): documented individually as generated artifacts tied to `logs.proto`.
- Built binaries (`authApp`, `brokerApp`, etc.): documented individually as build outputs.
- Runtime database/message-broker volume files (`project/db-data/**`): classified as generated runtime state; not manually maintained.

## Workspace Metadata

- `.gitignore`: ignore policy for editor files, env files, logs, local DB volumes, binaries, and temporary files.
- `.vscode/settings.json`: VS Code association so `*.gohtml` is treated as HTML for syntax highlighting.
- `.idea/Microservices.iml`: JetBrains module descriptor (IDE metadata).
- `.idea/copilot.data.migration.agent.xml`: JetBrains Copilot/Codex migration metadata.
- `.idea/copilot.data.migration.ask2agent.xml`: JetBrains Copilot/Codex migration metadata.
- `.idea/copilot.data.migration.edit.xml`: JetBrains Copilot/Codex migration metadata.
- `.idea/go.imports.xml`: JetBrains Go import settings metadata.
- `.idea/golang-microservices.iml`: JetBrains module descriptor for the workspace.
- `.idea/material_theme_project_new.xml`: JetBrains theme/plugin project metadata.
- `.idea/modules.xml`: JetBrains module index metadata.
- `.idea/vcs.xml`: JetBrains VCS integration metadata.
- `.idea/workspace.xml`: JetBrains per-workspace state (local IDE settings/history).

## `authentication-service/`

- `authentication-service/go.mod`: module definition and direct dependency declarations (bcrypt and DB/http stack via transitive deps).
- `authentication-service/go.sum`: dependency checksum lockfile for reproducible module resolution.
- `authentication-service/authentication-service.dockerfile`: minimal runtime image copying `authApp` into Alpine and executing it.
- `authentication-service/cmd/api/main.go`: service bootstrap, HTTP server startup, Postgres connection retry logic, and repository wiring helper.
- `authentication-service/cmd/api/routes.go`: chi router setup, CORS policy, heartbeat route, and `/authenticate` route registration.
- `authentication-service/cmd/api/helpers.go`: shared JSON read/write/error response utilities with request body size limiting.
- `authentication-service/cmd/api/handlers.go`: authenticate endpoint handler, credential validation flow, and login event forwarding to logger service.
- `authentication-service/cmd/api/setup_test.go`: test bootstrap that injects `PostgresTestRepository` into shared test config.
- `authentication-service/cmd/api/routes_test.go`: asserts expected routes exist in router configuration.
- `authentication-service/cmd/api/handlers_test.go`: handler-level test with custom HTTP transport to mock downstream logger call.
- `authentication-service/data/repository.go`: repository interface contract used by handlers and tests.
- `authentication-service/data/models.go`: Postgres repository implementation for CRUD, password hashing, and password verification.
- `authentication-service/data/test-models.go`: test repository implementation returning deterministic fixture-like user responses.
- `authentication-service/authApp`: compiled Linux ARM64 authentication binary (build artifact, not source).

## `broker-service/`

- `broker-service/go.mod`: broker module definition with chi/cors and messaging/grpc dependencies.
- `broker-service/go.sum`: dependency checksum lockfile.
- `broker-service/broker-service.dockerfile`: Alpine runtime image that copies and runs `brokerApp`.
- `broker-service/cmd/api/main.go`: broker bootstrap, RabbitMQ connection with exponential backoff, and HTTP server startup.
- `broker-service/cmd/api/routes.go`: route registration for broker entrypoint, submission handler, gRPC logging endpoint, and heartbeat.
- `broker-service/cmd/api/helpers.go`: JSON request/response helpers and consistent error payload formatting.
- `broker-service/cmd/api/handlers.go`: core orchestration logic for `auth`, `log`, and `mail` actions; includes HTTP, RPC, gRPC, and optional RabbitMQ logging paths.
- `broker-service/event/event.go`: RabbitMQ exchange/queue declaration helpers shared by consumer and emitter.
- `broker-service/event/emitter.go`: RabbitMQ publisher implementation for topic exchange events.
- `broker-service/event/consumer.go`: RabbitMQ consumer implementation and forwarding logic to logger HTTP endpoint.
- `broker-service/logs/logs.proto`: protobuf service contract for gRPC log write operations.
- `broker-service/logs/logs.pb.go`: generated protobuf message code for `Log`, `LogRequest`, `LogResponse`.
- `broker-service/logs/logs_grpc.pb.go`: generated gRPC client/server stubs for `LogService`.
- `broker-service/brokerApp`: compiled Linux ARM64 broker binary (build artifact, not source).

## `front-end/`

- `front-end/go.mod`: front-end module declaration.
- `front-end/main.go`: web server entrypoint, template rendering pipeline, and `BROKER_URL` template data injection.
- `front-end/templates/base.layout.gohtml`: base HTML layout with content and JS blocks.
- `front-end/templates/header.partial.gohtml`: document head metadata and Bootstrap CSS include.
- `front-end/templates/footer.partial.gohtml`: page footer markup.
- `front-end/templates/test.page.gohtml`: interactive test UI and JavaScript actions that call broker endpoints for auth/log/mail/grpc flows.
- `front-end/frontApp`: compiled Linux ARM64 front-end binary (build artifact, not source).

## `listener-service/`

- `listener-service/go.mod`: listener module declaration with RabbitMQ client dependency.
- `listener-service/go.sum`: dependency checksums.
- `listener-service/listener-service.dockerfile`: Alpine runtime image that copies and runs `listenerApp`.
- `listener-service/main.go`: listener bootstrap and RabbitMQ connection/retry setup; subscribes to log severity topics.
- `listener-service/event/event.go`: exchange and queue declaration helpers for RabbitMQ topic consumption.
- `listener-service/event/consumer.go`: long-running topic consumer; decodes payload and forwards log events to logger HTTP API.
- `listener-service/listenerApp`: compiled Linux ARM64 listener binary (build artifact, not source).

## `logger-service/`

- `logger-service/go.mod`: logger module with chi/cors, MongoDB driver, and gRPC dependencies.
- `logger-service/go.sum`: dependency checksum lockfile.
- `logger-service/logger-service.dockerfile`: Alpine runtime image that copies and runs `loggerServiceApp`.
- `logger-service/cmd/api/main.go`: logger bootstrap, Mongo connection, HTTP server setup, and background RPC/gRPC server startup.
- `logger-service/cmd/api/routes.go`: HTTP routing for `/log` and heartbeat endpoint.
- `logger-service/cmd/api/helpers.go`: JSON decoding/encoding helpers and standardized error responses.
- `logger-service/cmd/api/handlers.go`: HTTP handler accepting log entries and inserting into MongoDB.
- `logger-service/cmd/api/rpc.go`: net/rpc endpoint implementation (`LogInfo`) for broker-to-logger RPC logging.
- `logger-service/cmd/api/grpc.go`: gRPC server implementation for `logs.LogService/Write`.
- `logger-service/data/models.go`: Mongo-backed data layer for inserting, listing, fetching, dropping, and updating log entries.
- `logger-service/logs/logs.proto`: protobuf contract that defines gRPC log write request/response schema.
- `logger-service/logs/logs.pb.go`: generated protobuf message types.
- `logger-service/logs/logs_grpc.pb.go`: generated gRPC service stubs.
- `logger-service/loggerServiceApp`: compiled Linux ARM64 logger binary (build artifact, not source).

## `mail-service/`

- `mail-service/go.mod`: mail service module and mail/template-related dependencies.
- `mail-service/go.sum`: dependency checksums.
- `mail-service/mail-service.dockerfile`: Alpine runtime image; copies service binary and template directory.
- `mail-service/cmd/api/main.go`: mail service bootstrap and env-driven SMTP configuration initialization.
- `mail-service/cmd/api/routes.go`: route registration for `/send` and heartbeat.
- `mail-service/cmd/api/helpers.go`: JSON request/response helpers and error response shaping.
- `mail-service/cmd/api/handlers.go`: `/send` handler that parses request payload and dispatches SMTP send.
- `mail-service/cmd/api/mailer.go`: SMTP client abstraction, HTML/plain template rendering, CSS inlining, encryption mode mapping, and send logic.
- `mail-service/templates/mail.html.gohtml`: HTML email body template.
- `mail-service/templates/mail.plain.gohtml`: plain-text email body template.
- `mail-service/mailerApp`: compiled Linux ARM64 mail service binary (build artifact, not source).

## `project/` Infrastructure

- `project/Makefile`: local automation for building service binaries and starting/stopping Docker Compose stack.
- `project/docker-compose.yaml`: local multi-container topology for broker, auth, logger, mail, listener, and supporting infra (Postgres, Mongo, RabbitMQ, MailHog).
- `project/postgres.yml`: standalone Postgres compose definition.
- `project/ingress.yml`: Kubernetes ingress routing rules for front-end and broker hostnames.
- `project/k8s/authentication.yml`: authentication deployment/service manifest including `DSN` env setup.
- `project/k8s/broker.yml`: broker deployment/service manifest.
- `project/k8s/front-end.yml`: front-end deployment/service manifest, including `BROKER_URL`.
- `project/k8s/listener.yml`: listener deployment/service manifest.
- `project/k8s/logger.yml`: logger deployment/service manifest, exposing HTTP, RPC, and gRPC ports.
- `project/k8s/mail.yml`: mailer deployment/service manifest with SMTP environment variables.
- `project/k8s/mailhog.yml`: mail capture UI/SMTP deployment and service.
- `project/k8s/mongo.yml`: MongoDB deployment/service for logger persistence.
- `project/k8s/rabbit.yml`: RabbitMQ deployment/service for event messaging.

## Generated Runtime State

- `project/db-data/**`: persisted Postgres, MongoDB, and RabbitMQ runtime volume data (1649 files at scan time); generated by containers, not manually authored source.
