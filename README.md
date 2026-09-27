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
- `GET /api/v1/state` — household, inventory, candidate meals, history, and nullable `selectedMealId`
- `POST /api/v1/meals/{id}/confirm` — confirm a meal with JSON `{"rating":"loved|okay|not-for-us","note":"optional text"}`

A successful confirmation returns the complete updated state. Inventory deductions, the new first history entry, and selection clearing happen atomically; inventory quantities cannot fall below zero.

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
