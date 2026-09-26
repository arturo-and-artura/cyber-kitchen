---
domain: Architecture
status: Draft
entry_points: []
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

One `cyber-kitchen` binary exposes two interfaces:

- `cyber-kitchen serve` runs the versioned HTTP API.
- `cyber-kitchen admin <command>` runs maintenance operations such as migrations, inspection, repair, or backfills.

Product workflows remain API-only. HTTP handlers and administration commands translate inputs and outputs while shared application and domain services own business behavior.

## Package Boundaries

The first implementation will use Go and the following package boundaries:

| Path | Responsibility |
|------|----------------|
| `cmd/cyber-kitchen` | Binary entry point and command wiring |
| `internal/api` | HTTP transport and versioned contracts |
| `internal/admin` | Operational command handlers |
| `internal/domain` | Product rules and entities |
| `internal/store` | Persistence ports and adapters |
| `internal/ai` | AI provider ports, validation, and fallbacks |

Dependencies point inward toward domain behavior. Persistence engines, external providers, and command frameworks remain adapters rather than domain dependencies.

## Invariants

- Product behavior MUST be available through versioned HTTP APIs rather than a product CLI or TUI.
- HTTP and administration entry points MUST NOT own business rules.
- Administration commands MUST be explicit, auditable, and protected by the same authorization boundaries as equivalent service operations.
- External AI output MUST be validated before it can affect allergies, inventory, meal history, or other persistent state.
- Client presentation code MUST remain outside the backend repository.
