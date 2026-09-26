# Cyber Kitchen Backend

The backend service for Cyber Kitchen. This repository owns the product APIs, domain logic, persistence, background work, and operational administration commands.

The browser client lives in [`arturo-and-artura/cyber-kitchen-web`](https://github.com/arturo-and-artura/cyber-kitchen-web). Future native clients should use their own suffixed repositories and consume the same versioned API.

## Direction

- **Language:** Go
- **Delivery:** one statically linked `cyber-kitchen` binary
- **Runtime entry point:** `cyber-kitchen serve`
- **Operations:** `cyber-kitchen admin <command>`
- **Product surface:** HTTP APIs only; product features do not need a CLI or TUI

Go is the default because this service values maintainability, predictable code generation and review, fast builds, straightforward concurrency, and simple single-binary deployment more than Rust's additional low-level control. See [`docs/backend-architecture.md`](docs/backend-architecture.md) for the decision and initial boundaries.

## Status

The repository organization and backend direction are established. The first implementation PR will add the Go module, API skeleton, configuration, health checks, and automated verification.

## Documentation

- [`docs/backend-architecture.md`](docs/backend-architecture.md) — language decision, binary shape, and initial package boundaries
- [`.aidoc/INDEX.md`](.aidoc/INDEX.md) — durable documentation index
- [`AGENT.md`](AGENT.md) — implementation and delivery guidance
