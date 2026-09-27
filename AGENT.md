# Cyber Kitchen repository guide

## Purpose

Cyber Kitchen is a responsive household meal decision and cooking prototype.

## Architecture

- React + TypeScript + Vite single-page app; no router or backend in this MVP.
- `App.tsx` owns the view state and persisted demo state.
- Screen components live in `src/components`; fixtures in `src/data`; deterministic state helpers in `src/lib`.
- Keep recommendation explanations explicit. Do not imply that the current recommendations use a live AI service.
- Keep allergy constraints visible at decision points and never mock a recommendation that violates them.
- Persist demo data under one versioned local-storage key. Inventory consumption must be deterministic and previewed before commit.

## Design and accessibility

- Preserve semantic headings, labels, keyboard focus states, sufficient contrast, and reduced-motion behavior.
- Validate both compact mobile and desktop layouts when changing a screen.
- Prefer plain-language actions and explain why a meal is suggested.

## Commands

```bash
npm run dev
npm test
npm run typecheck
npm run lint
npm run build
```

## Definition of done

Changes must keep the primary Today → Choose → Cook → Confirm → Today loop runnable; update synchronized product documentation when states or data change; and pass test, typecheck, lint, and build. For meaningful UI changes, inspect the app at desktop and mobile widths.

## Delivery policy

Work on feature branches and open a PR to the protected default branch. Do not deploy, change production data, add a real external service, or introduce credentials without explicit approval.
