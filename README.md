# Cyber Kitchen Backend

The Go backend for Cyber Kitchen. It serves the household state and owns inventory and meal-history updates independently of any client UI.

The browser client lives in [`arturo-and-artura/cyber-kitchen-web`](https://github.com/arturo-and-artura/cyber-kitchen-web).

## Run locally

Go 1.24 or newer is required.

```sh
go run ./cmd/cyber-kitchen serve
```

The service listens on `:8080`, allows `http://localhost:5173`, and persists state to `data/cyber-kitchen.json` by default. Override these settings with flags or environment variables:

```sh
go run ./cmd/cyber-kitchen serve -listen=:9090 -cors-origins=http://localhost:3000,http://localhost:5173 -data-file=/tmp/cyber-kitchen.json
# or CYBER_KITCHEN_LISTEN=:9090 CYBER_KITCHEN_CORS_ORIGINS=http://localhost:3000 CYBER_KITCHEN_DATA_FILE=/tmp/cyber-kitchen.json
```

The data file is created from the MVP seed on first launch. Successful meal confirmations are written atomically and remain available after a service restart.

## API

- `GET /healthz` — service health
- `GET /api/v1/household` — raw household resource with `name`, `members`, `constraints`, and `goals`
- `GET /api/v1/inventory` — inventory collection in `{"inventory": [...]}`
- `GET /api/v1/meals` — candidate meals and nullable selection in `{"meals": [...], "selectedMealId": null}`
- `GET /api/v1/history` — meal history in `{"history": [...]}`
- `POST /api/v1/meals/{id}/confirm` — confirm a meal with JSON `{"rating":"loved|okay|not-for-us","note":"optional text"}`

Clients assemble application state from the four resource reads; the service does not expose an aggregate state endpoint. A successful confirmation returns only the committed `inventory`, `history`, and nullable `selectedMealId`. Inventory deductions, the new first history entry, and selection clearing happen atomically; inventory quantities cannot fall below zero. See the [backend architecture](.aidoc/architecture/backend.md#implemented-http-contract) for field-level contracts.

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
