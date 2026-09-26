# Cyber Kitchen

A polished, UX-first prototype for a shared household meal-planning loop. It turns visible household constraints and a mocked inventory into three explainable dinner recommendations, guides cooking, collects feedback, and deterministically updates inventory.

> This MVP is entirely local: recommendations and data are mocked, there are no API calls, and demo state is persisted in `localStorage`.

## Run locally

Prerequisite: Node.js 22+ and npm.

```bash
npm install
npm run dev
```

Open the URL Vite prints (normally <http://localhost:5173>).

## Build and preview

Build the portable static site, then serve it on any host and port:

```bash
npm run build
npm run preview -- --host 0.0.0.0 --port 4173
```

The build output is written to `dist/` and does not depend on a Cyber Kitchen-specific hostname. It can be served by any static host or reverse proxy. For direct links to work, configure the host to fall back to `index.html` for unknown paths; this MVP currently keeps navigation in client state and makes no backend requests.

## Quality commands

```bash
npm test          # Vitest + Testing Library
npm run typecheck # strict TypeScript
npm run lint      # ESLint
npm run build     # production build
```

## Demo script

1. On **Today**, notice that peanut-free, dairy-light, high-protein, and food-waste goals are visible.
2. Select **Choose tonight’s meal** and compare three recommendations and their explanations.
3. Choose a meal, check off ingredients, and advance through all cooking steps.
4. Finish cooking, give feedback, and review the exact inventory adjustment preview.
5. Confirm. The completion is reflected on Today, in Inventory, and in History.

Use the main navigation to explore **Inventory** (search/filter/low-stock states) and **History**. To reset the demo, remove the `cyber-kitchen-demo-v1` key in browser local storage.

## Documentation

- [`docs/mvp-user-flow-and-screen-map.md`](docs/mvp-user-flow-and-screen-map.md) — implemented flow, screen states, and data model
- [`docs/screenshots/today-desktop.png`](docs/screenshots/today-desktop.png) and [`today-mobile.png`](docs/screenshots/today-mobile.png) — browser verification captures
- [`.aidoc/INDEX.md`](.aidoc/INDEX.md) — durable project documentation index
- [`AGENT.md`](AGENT.md) — contribution and verification guidance
