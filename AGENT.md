# Cyber Kitchen backend repository guide

## Purpose

Cyber Kitchen is the backend service for the household meal decision and cooking product.

## Architecture

- Build one deployable `cyber-kitchen` binary.
- Expose product capabilities through versioned HTTP APIs invoked by separate client repositories.
- Put runtime startup under `cyber-kitchen serve` and operational maintenance under `cyber-kitchen admin <command>`; do not build product-facing CLI or TUI flows.
- Keep domain rules independent from HTTP, CLI, storage, and external AI providers.
- Keep admin commands explicit, auditable, safe to retry where practical, and protected by the same authorization boundaries as equivalent service operations.
- Treat AI output as untrusted input. Validate it before it can affect allergies, inventory, meal history, or other persistent state.

## Package layout

When implementation is added, use these boundaries unless the architecture documentation is updated first:

- `cmd/cyber-kitchen` — binary entry point and command wiring
- `internal/api` — HTTP transport and versioned contracts
- `internal/admin` — operational command handlers
- `internal/domain` — product rules and entities
- `internal/store` — persistence adapters
- `internal/ai` — AI provider boundaries and validation

## Definition of done

Changes must include focused tests for domain behavior, keep API and admin entry points thin, update relevant architecture documentation, and pass formatting, static analysis, tests, and build verification. Add integration tests when a change crosses transport or persistence boundaries.

## Delivery policy

Work on feature branches and open a PR to the protected default branch. Do not deploy, change production data, add a real external service, or introduce credentials without explicit approval.
