# Cyber Kitchen backend repository guide

## Working instructions

- Read [the documentation index](.aidoc/INDEX.md) and [backend architecture](.aidoc/architecture/backend.md) before changing service interfaces, package boundaries, persistence, administration commands, or AI integration.
- Treat the architecture document as canonical for the implementation stack, runtime interfaces, package boundaries, and cross-cutting invariants. Update it before making an intentional architecture change.
- Keep domain behavior independent from HTTP, CLI, storage, and external AI providers.
- Keep API and administration entry points thin, and validate untrusted external output before it can affect household constraints or persistent state.

## Definition of done

Changes must include focused tests for domain behavior, update relevant architecture documentation, and pass formatting, static analysis, tests, and build verification. Add integration tests when a change crosses transport or persistence boundaries.

## Delivery policy

Work on feature branches and open a PR to the protected default branch. Do not deploy, change production data, add a real external service, or introduce credentials without explicit approval.
