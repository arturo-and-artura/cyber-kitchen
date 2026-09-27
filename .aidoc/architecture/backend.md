---
domain: Architecture
status: Active
entry_points:
  - cmd/cyber-kitchen/main.go
  - internal/api/server.go
  - internal/domain/service.go
dependencies:
  - ../INDEX.md
---

# Backend architecture

Cyber Kitchen uses a single backend service to keep household meal rules consistent across clients. This document defines the intended service boundaries that implementation must preserve as the backend is introduced.

## Related Docs

| Document | Relationship |
|----------|-------------|
| [Documentation index](../INDEX.md) | Canonical documentation entry point |
| [Repository guide](../../AGENT.md) | Active implementation and delivery rules |

## Why the Backend Exists

Cyber Kitchen needs one authoritative place for product APIs, domain rules, persistence, background work, and operational administration. Keeping those responsibilities behind a versioned HTTP boundary lets browser and future clients share behavior without sharing deployable source.

The backend boundary protects household constraints and persistent state from transport details and untrusted external output. Allergy rules, inventory changes, and meal history remain domain concerns regardless of which client or provider initiates a request.

## What the Backend Owns

The `cyber-kitchen` repository owns the server and operational administration. Client repositories own their presentation code and consume versioned HTTP APIs.

One `cyber-kitchen` binary owns the runtime interfaces:

- `cyber-kitchen serve` currently runs the versioned HTTP API.
- `cyber-kitchen admin <command>` is reserved for future maintenance operations such as migrations, inspection, repair, or backfills.

Product workflows remain API-only. HTTP handlers—and future administration commands—translate inputs and outputs while shared application and domain services own business behavior.

## Package Boundaries

The implementation uses Go and the following package boundaries:

| Path | Responsibility |
|------|----------------|
| `cmd/cyber-kitchen` | Binary entry point and command wiring |
| `internal/api` | HTTP transport and versioned contracts |
| `internal/admin` | Future operational command handlers |
| `internal/domain` | Product rules and entities |
| `internal/store` | Persistence ports and adapters; the service uses an atomic file-backed adapter and retains an in-memory adapter for focused tests |
| `internal/seed` | Initial household, inventory, meal, and history fixture data |
| `internal/ai` | Future AI provider ports, validation, and fallbacks |

## Implemented HTTP Contract

The service currently exposes JSON over these routes:

| Method and path | Stable response contract |
|-----------------|--------------------------|
| `GET /healthz` | `{"status":"ok"}` |
| `GET /api/v1/household` | `{"name":string,"members":[...],"constraints":[string],"goals":[string]}` (the household resource directly, without an envelope) |
| `GET /api/v1/inventory` | `{"inventory":[{"id":string,"name":string,"amount":number,"unit":string,"category":string,"lowAt":number}]}` |
| `GET /api/v1/meals` | `{"meals":[...],"selectedMealId":string|null}`; each meal includes its display metadata, tags, ingredients, and steps. |
| `GET /api/v1/history` | `{"history":[{"id":string,"mealId":string,"mealName":string,"emoji":string,"cookedAt":string,"rating":string,"note":string}]}`; `cookedAt` is UTC RFC 3339 with millisecond precision. |
| `POST /api/v1/meals/{id}/confirm` | Accepts `{"rating":"loved|okay|not-for-us","note":"..."}` and returns only `{"inventory":[...],"history":[...],"selectedMealId":string|null}` after commit. |

All collection fields are JSON arrays, including when empty. Clients assemble application state from the four resource reads; there is no aggregate state endpoint. Meal confirmation is one atomic store operation. It finds the meal by path ID, subtracts each recipe ingredient from the corresponding inventory item without allowing a negative amount, prepends a timestamped history record, and clears `selectedMealId`. The success response is built from the committed state returned by that operation. Invalid ratings and unknown meals leave state unchanged. Unknown JSON fields are rejected, and transport errors retain the `{"error":string}` contract.

The initial state mirrors the browser client's MVP fixture; the household profile comes from the same client's household card. On first launch, the file-backed store writes that seed to the configured data file. Every successful domain update is encoded to a protected temporary file and atomically renamed over the active file before the new in-memory snapshot becomes visible. A rejected or failed write leaves both the published snapshot and the prior data file unchanged. A later service process loads the committed file rather than reseeding.

The persisted file is an internal adapter format, not a public contract. During development it may change together with the current code and data; migration or compatibility behavior belongs only to a future production persistence design.

## Runtime Configuration

`cyber-kitchen serve` accepts `-listen`, `-cors-origins`, and `-data-file`. Their environment equivalents are `CYBER_KITCHEN_LISTEN` (default `:8080`), `CYBER_KITCHEN_CORS_ORIGINS` (default `http://localhost:5173`), and `CYBER_KITCHEN_DATA_FILE` (default `data/cyber-kitchen.json`). The origins value is a comma-separated allowlist; `*` enables a wildcard response. CORS applies at the HTTP adapter only.

Dependencies point inward toward domain behavior. Persistence engines, external providers, and command frameworks remain adapters rather than domain dependencies.

## Invariants

- Product behavior MUST be available through versioned HTTP APIs rather than a product CLI or TUI.
- HTTP and administration entry points MUST NOT own business rules.
- Administration commands MUST be explicit, auditable, and protected by the same authorization boundaries as equivalent service operations.
- External AI output MUST be validated before it can affect allergies, inventory, meal history, or other persistent state.
- Client presentation code MUST remain outside the backend repository.
