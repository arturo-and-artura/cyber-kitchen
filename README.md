# Cyber Kitchen Backend

The backend service for Cyber Kitchen. This repository owns product APIs, domain logic, persistence, background work, and operational administration commands.

The browser client lives in [`arturo-and-artura/cyber-kitchen-web`](https://github.com/arturo-and-artura/cyber-kitchen-web). Other clients consume the same versioned HTTP API from their own repositories.

## Current state

The backend implementation has not started. The established implementation uses Go and one `cyber-kitchen` binary with an HTTP service entry point and a separate administration command family. Product workflows remain API-only.

## Documentation

- [Backend architecture](.aidoc/architecture/backend.md) — service boundary, interfaces, package boundaries, and safety constraints
- [Documentation index](.aidoc/INDEX.md) — canonical reading paths
- [Repository guide](AGENT.md) — implementation and delivery rules
