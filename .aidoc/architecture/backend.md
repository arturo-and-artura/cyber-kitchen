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
| `internal/store` | Persistence ports and adapters; currently an atomic in-memory adapter |
| `internal/seed` | Initial household, inventory, meal, and history fixture data |
| `internal/ai` | Future AI provider ports, validation, and fallbacks |

## Implemented HTTP Contract

The service currently exposes JSON over these routes:

| Method and path | Behavior |
|-----------------|----------|
| `GET /healthz` | Returns `{"status":"ok"}` for process health checks. |
| `GET /api/v1/state` | Returns the household profile, inventory, candidate meals, meal history, and nullable `selectedMealId`. |
| `POST /api/v1/meals/{id}/confirm` | Accepts `{"rating":"loved|okay|not-for-us","note":"..."}` and returns the complete updated state. |

Meal confirmation is one atomic store operation. It finds the meal by path ID, subtracts each recipe ingredient from the corresponding inventory item without allowing a negative amount, prepends a timestamped history record, and clears `selectedMealId`. Invalid ratings and unknown meals leave state unchanged. Unknown JSON fields are rejected.

The initial state mirrors `cyber-kitchen-web/src/data/mockData.ts`; the household profile comes from the same client's household card. State is process-local and resets when the service restarts.

## Runtime Configuration

`cyber-kitchen serve` accepts `-listen` and `-cors-origins`. Their environment equivalents are `CYBER_KITCHEN_LISTEN` (default `:8080`) and `CYBER_KITCHEN_CORS_ORIGINS` (default `http://localhost:5173`). The origins value is a comma-separated allowlist; `*` enables a wildcard response. CORS applies at the HTTP adapter only.

Dependencies point inward toward domain behavior. Persistence engines, external providers, and command frameworks remain adapters rather than domain dependencies.

## Invariants

- Product behavior MUST be available through versioned HTTP APIs rather than a product CLI or TUI.
- HTTP and administration entry points MUST NOT own business rules.
- Administration commands MUST be explicit, auditable, and protected by the same authorization boundaries as equivalent service operations.
- External AI output MUST be validated before it can affect allergies, inventory, meal history, or other persistent state.
- Client presentation code MUST remain outside the backend repository.
