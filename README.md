# Cyber Kitchen Backend

The Go backend for Cyber Kitchen. It serves the household state and owns inventory and meal-history updates independently of any client UI.

The browser client lives in [`arturo-and-artura/cyber-kitchen-web`](https://github.com/arturo-and-artura/cyber-kitchen-web).

## Run locally

Go 1.24 or newer is required.

```sh
go run ./cmd/cyber-kitchen serve
```

The service listens on `:8080` and allows `http://localhost:5173` by default. Override either setting with flags or environment variables:

```sh
go run ./cmd/cyber-kitchen serve -listen=:9090 -cors-origins=http://localhost:3000,http://localhost:5173
# or CYBER_KITCHEN_LISTEN=:9090 CYBER_KITCHEN_CORS_ORIGINS=http://localhost:3000
```

State is currently in memory and starts from the frontend MVP fixture on each process launch.

## API

- `GET /healthz` — service health
- `GET /api/v1/household` — household profile in `{"household": ...}`
- `GET /api/v1/inventory` — inventory collection in `{"inventory": [...]}`
- `GET /api/v1/meals` — candidate meals and nullable selection in `{"meals": [...], "selectedMealId": null}`
- `GET /api/v1/history` — meal history in `{"history": [...]}`
- `GET /api/v1/state` — temporary compatibility aggregate for existing clients
- `POST /api/v1/meals/{id}/confirm` — confirm a meal with JSON `{"rating":"loved|okay|not-for-us","note":"optional text"}`

A successful confirmation returns only the committed `inventory`, `history`, and nullable `selectedMealId`. Inventory deductions, the new first history entry, and selection clearing happen atomically; inventory quantities cannot fall below zero. See the [backend architecture](.aidoc/architecture/backend.md#implemented-http-contract) for field-level contracts.

## Verify

```sh
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/cyber-kitchen
```

## Documentation

- [Backend architecture](.aidoc/architecture/backend.md) — service boundary, API contract, package boundaries, and runtime configuration
- [Documentation index](.aidoc/INDEX.md) — canonical reading paths
- [Repository guide](AGENT.md) — implementation and delivery rules
