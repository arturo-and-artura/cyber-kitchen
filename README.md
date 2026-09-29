# Cyber Kitchen Backend

The Go service for Cyber Kitchen's single-household kitchen-assistance agent. It owns PostgreSQL state, assembles focused recommendation context, validates the predefined agent response, and keeps inventory/history mutations deterministic and user-confirmed.

The browser client lives in [`arturo-and-artura/cyber-kitchen-web`](https://github.com/arturo-and-artura/cyber-kitchen-web).

## Run locally

Go 1.24+, Docker, and Docker Compose are required.

```sh
docker compose up -d --wait
DEEPSEEK_API_KEY='<runtime secret>' go run ./cmd/cyber-kitchen serve
```

The service listens on `:8080`, allows `http://localhost:5173`, and uses the Compose PostgreSQL database by default. Override these with `CYBER_KITCHEN_LISTEN`, `CYBER_KITCHEN_CORS_ORIGINS`, and `CYBER_KITCHEN_DATABASE_URL`, or the corresponding `-listen`, `-cors-origins`, and `-database-url` flags.

The development schema is disposable and intentionally has no migration/compatibility layer. Reset it with `docker compose down --volumes`.

### DeepSeek recommendations

The service requires a DeepSeek API key at startup:

```sh
DEEPSEEK_API_KEY='<runtime secret>' \
go run ./cmd/cyber-kitchen serve
```

Missing configuration fails startup; recommendation-provider or validation failures return an error without replacing the current meals. `DEEPSEEK_MODEL` defaults to `deepseek-chat`. `DEEPSEEK_BASE_URL` may select a compatible self-hosted endpoint. The key is read only from the process environment; it is not stored, logged, or added to prompts. Do not commit keys or local secret-file paths.

## API

Resource reads are available at `/api/v1/household`, `/api/v1/inventory`, `/api/v1/meals`, and `/api/v1/history`. The pilot also supports:

- `PUT /api/v1/household` — replace constraints and goals
- `PUT /api/v1/inventory/{id}` — create or replace an inventory item
- `DELETE /api/v1/inventory/{id}` — delete an inventory item
- `POST /api/v1/recommendations/generate` — run one focused kitchen-agent turn
- `POST /api/v1/meals/{id}/confirm` — atomically commit the explicit meal confirmation

See the [backend architecture](.aidoc/architecture/backend.md#http-contract) for boundaries and safety invariants.

## Verify

```sh
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/cyber-kitchen
CYBER_KITCHEN_TEST_DATABASE_URL='postgres://cyber_kitchen:cyber_kitchen@localhost:5432/cyber_kitchen?sslmode=disable' go test ./internal/store -run TestPostgres
```

## Documentation

- [Backend architecture](.aidoc/architecture/backend.md)
- [Documentation index](.aidoc/INDEX.md)
- [Repository guide](AGENT.md)
