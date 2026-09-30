# Cyber Kitchen Backend

The Go service for Cyber Kitchen's single-household kitchen-assistance agent. It owns PostgreSQL state, assembles focused recommendation context, validates the predefined agent response, and keeps inventory/history mutations deterministic and user-confirmed.

The browser client lives in [`arturo-and-artura/cyber-kitchen-web`](https://github.com/arturo-and-artura/cyber-kitchen-web).

## Run locally

Go 1.24+, Docker, and Docker Compose are required.

```sh
docker compose up -d --wait
./scripts/serve
```

The service listens on `:8080`, allows `http://localhost:5173`, and uses the Compose PostgreSQL database by default. Override these with `CYBER_KITCHEN_LISTEN`, `CYBER_KITCHEN_CORS_ORIGINS`, and `CYBER_KITCHEN_DATABASE_URL`, or the corresponding `-listen`, `-cors-origins`, and `-database-url` flags. Additional server flags may be passed to `./scripts/serve`.

The development schema is disposable and intentionally has no migration/compatibility layer. Reset it with `docker compose down --volumes`.

### DeepSeek recommendations

Store the DeepSeek API key in the ignored private file used by the development script:

```sh
mkdir -p .secrets
chmod 700 .secrets
${EDITOR:-vi} .secrets/deepseek-api-key
chmod 600 .secrets/deepseek-api-key
./scripts/serve
```

The app remains explorable when AI recommendations are not configured. Household, inventory, meal, history, and health resources work normally; the recommendation endpoint returns a stable error code and a user-facing explanation that the frontend can present. Provider or validation failures return retry guidance without replacing the current meals. `DEEPSEEK_MODEL` defaults to `deepseek-chat`. `DEEPSEEK_BASE_URL` may select a compatible self-hosted endpoint. The key file path may be changed with `-deepseek-api-key-file`; its contents are never logged, persisted, or added to prompts.

## API

Resource reads are available at `/api/v1/household`, `/api/v1/inventory`, `/api/v1/meals`, and `/api/v1/history`. The pilot also supports:

- `PUT /api/v1/household` — atomically replace members, constraints, goals, and preferences (the household name remains server-owned)
- `PUT /api/v1/inventory/{id}` — create or replace an inventory item, including optional count, storage, recorded-date, and notes metadata
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
