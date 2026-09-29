---
domain: Architecture
status: Active
entry_points:
  - cmd/cyber-kitchen/main.go
  - internal/api/server.go
  - internal/agent/agent.go
  - internal/domain/service.go
dependencies:
  - ../INDEX.md
---

# Backend architecture

Cyber Kitchen is a focused kitchen-assistance agent backed by deterministic application services. The application assembles household state for each agent turn, accepts only a predefined validated response, and keeps every persistent mutation outside model authority.

## Related Docs

| Document | Relationship |
|----------|-------------|
| [Documentation index](../INDEX.md) | Canonical documentation entry point |
| [Repository guide](../../AGENT.md) | Active implementation and delivery rules |

## Why the Backend Exists

One backend owns the single-household pilot's state, kitchen-agent turn, safety validation, and mutations. Browser clients consume versioned HTTP resources; they do not duplicate persistence or recommendation policy.

Cooking has consequences. Inventory and history changes therefore remain deterministic domain operations that occur only after explicit user confirmation. Agent output may replace recommendations only after structural and inventory-safety validation.

## Focused Agent Runtime

`internal/agent.Runner` is the runtime boundary for one stateless kitchen-assistance turn:

1. `BuildPrompt` assembles household constraints and goals, current inventory, and history from application-owned state.
2. A narrow `Model` implementation may produce one JSON document in the predefined `Response` format.
3. `Validate` requires exactly three complete, uniquely identified meals, known inventory references, available quantities, supported difficulty values, and non-empty cooking steps.
4. Provider failures, invalid JSON, unknown fields, or unsafe content fail the turn without changing the current recommendations.
5. Validated recommendations are persisted atomically. The model cannot mutate inventory, history, household state, or files and has no tools or persistent session.

The pilot uses one direct DeepSeek adapter for recommendation generation. It implements the stateless `Model` boundary, sends one prompt to the OpenAI-compatible chat-completions API, requests a JSON object, and returns only the assistant content to the strict validator. The runtime reads the key from a private file selected by `-deepseek-api-key-file`; the key remains outside command arguments, process-environment configuration, prompts, logs, and persistence.

This direct adapter fits the single-turn response contract without Pi RPC's subprocess supervision, sessions, compaction, JSONL event lifecycle, or tool lockdown. Pi should be reconsidered only when a concrete multi-step kitchen workflow benefits from those runtime capabilities. Additional providers should implement `Model`; provider selection must not leak into domain or HTTP packages.

## Package Boundaries

| Path | Responsibility |
|------|----------------|
| `cmd/cyber-kitchen` | Binary and runtime wiring |
| `internal/api` | HTTP transport and versioned contracts |
| `internal/agent` | Kitchen context/prompt assembly and strict response validation |
| `internal/domain` | Household, inventory, confirmation, and history rules |
| `internal/store` | PostgreSQL aggregate persistence and in-memory focused-test adapter |
| `internal/seed` | Disposable development fixture |

Dependencies point inward. The domain does not depend on HTTP, PostgreSQL, or model providers.

## HTTP Contract

| Method and path | Behavior |
|-----------------|----------|
| `GET /healthz` | Service health |
| `GET /api/v1/household` | Read the household profile |
| `PUT /api/v1/household` | Replace constraints and goals |
| `GET /api/v1/inventory` | Read inventory |
| `PUT /api/v1/inventory/{id}` | Create or replace one validated item |
| `DELETE /api/v1/inventory/{id}` | Delete one item |
| `GET /api/v1/meals` | Read current recommendations |
| `POST /api/v1/recommendations/generate` | Run one focused agent turn; unavailable AI capability and provider failures return stable error codes with user-facing messages and leave current recommendations unchanged |
| `GET /api/v1/history` | Read meal history |
| `POST /api/v1/meals/{id}/confirm` | Confirm rating/note, deduct inventory, prepend history, and clear selection atomically |

Unknown request fields are rejected. Collection fields are always arrays. Invalid input and unknown resources leave state unchanged.

## PostgreSQL Development Runtime

`compose.yaml` starts PostgreSQL 17 on loopback with a named local volume. `CYBER_KITCHEN_DATABASE_URL` (or `-database-url`) configures the connection; the default is the Compose development database. `CYBER_KITCHEN_LISTEN` and `CYBER_KITCHEN_CORS_ORIGINS` retain their existing meanings.

`-deepseek-api-key-file` defaults to `.secrets/deepseek-api-key`. An absent or empty file disables AI-backed recommendation generation while household, inventory, meal, history, and health resources remain available. The generation endpoint returns `ai_recommendations_unavailable` with a friendly explanation when that capability is disabled, and `ai_recommendation_failed` with retry guidance after provider or validation failures. `DEEPSEEK_MODEL` defaults to `deepseek-chat`, and `DEEPSEEK_BASE_URL` defaults to the provider's public API origin. The base URL override exists for focused integration testing and compatible self-hosted endpoints.

The disposable pilot schema stores the complete single-household aggregate as JSONB in one singleton row. Each accepted domain change writes the complete next state in one transaction before publishing it in memory. This makes confirmation atomic and restart-safe without migration or compatibility scaffolding. During development, schema changes may require `docker compose down --volumes` and a clean seed.

## Invariants

- The pilot supports one household and no authentication or multi-user behavior.
- External model output never directly mutates household, inventory, or history state.
- Recommendation output is strictly parsed and validated before persistence.
- Confirmation is the only cooking action that changes inventory/history, and inventory never falls below zero.
- No general-purpose model tools, filesystem capabilities, or persistent agent sessions are exposed.
- Production deployment and backward-compatibility machinery remain absent until explicitly required.
