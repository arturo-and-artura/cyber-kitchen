# Backend architecture direction

## Repository boundary

`cyber-kitchen` owns the backend server and operational administration. Web and future native clients live in separate repositories:

- `cyber-kitchen-web` — browser client
- future suffixed repositories — iOS, Android, or other clients

Clients consume versioned HTTP APIs and do not share deployable source with the server.

## Language decision: Go

Use Go for the first backend version.

Go is the more practical fit for this product and for AI-assisted implementation because it has a small language surface, conventional project structure, fast feedback, strong standard tooling, and straightforward single-binary deployment. Generated changes are generally easier to review, test, and keep idiomatic. Its garbage collector and runtime are acceptable for an API and administration service.

Rust would be preferable if profiling later demonstrates strict latency, memory, embedded, or systems-safety requirements that Go cannot meet. Choosing it now would add ownership, lifetime, async, and ecosystem complexity without a demonstrated product benefit, and would make AI-generated code more expensive to review safely.

## One binary, two interfaces

Build one `cyber-kitchen` binary with two top-level command families:

```text
cyber-kitchen serve
cyber-kitchen admin <command>
```

`serve` runs the versioned HTTP API. `admin` contains maintenance operations such as migrations, repair, inspection, or backfills as those needs appear. Product workflows remain API-only; there is no product CLI or TUI.

Both interfaces call the same application and domain services. HTTP handlers and CLI commands translate inputs and outputs but do not own business rules.

## Initial package boundaries

```text
cmd/cyber-kitchen  binary and command wiring
internal/api       HTTP transport and versioned contracts
internal/admin     operational command handlers
internal/domain    product rules and entities
internal/store     persistence ports and adapters
internal/ai        AI provider ports, validation, and fallbacks
```

Keep dependencies pointing inward toward domain behavior. External providers, persistence engines, and command frameworks are implementation details selected by later focused changes.
