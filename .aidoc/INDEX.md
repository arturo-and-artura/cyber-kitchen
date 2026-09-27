---
domain: Architecture
status: Active
entry_points: []
dependencies:
  - architecture/backend.md
---

# Cyber Kitchen backend documentation

This index is the canonical entry point for documentation about the backend product and its implementation boundaries. Read the architecture document before changing service interfaces or adding the initial code structure.

## Documentation map

| Document | Purpose |
|----------|---------|
| [Backend architecture](architecture/backend.md) | Defines repository ownership, runtime interfaces, package boundaries, and cross-cutting safety constraints. |
| [Repository guide](../AGENT.md) | Defines active implementation and delivery instructions. |
| [README](../README.md) | Summarizes repository purpose and current implementation state. |

## Reading chains

- **Begin backend implementation:** [Backend architecture](architecture/backend.md) → [Repository guide](../AGENT.md)
- **Change an API, command, storage, or AI boundary:** [Backend architecture](architecture/backend.md) → relevant implementation code
