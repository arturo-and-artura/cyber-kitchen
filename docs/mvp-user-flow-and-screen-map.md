# MVP user flow and screen map

Status: synchronized with the implementation in `src/` on 2026-09-26.

## Product promise

Cyber Kitchen helps a household answer “what should we make tonight?” without hiding the reasoning. It considers mocked inventory, household safety constraints, goals, time, and past feedback. The MVP deliberately simulates recommendation intelligence; it performs no network requests and makes no AI claims beyond the labeled mocked experience.

## Primary user flow

```text
Today
  └─ Choose tonight’s meal
       └─ Choose a Meal (3 explainable options)
            └─ Cook this meal
                 └─ Cook (ingredient checklist + sequential steps)
                      └─ Finish cooking
                           └─ Meal confirmation
                                ├─ Back to recipe
                                └─ Confirm meal & update inventory
                                     └─ Today (success state)
                                          ├─ Inventory reflects consumption
                                          └─ History includes feedback
```

Navigating to Today, Inventory, or History through the shell is always available. During the cooking loop, Today remains highlighted as the parent navigation area.

## Screen map

### 1. Today (`view = today`)

**Purpose:** orient the household and make the next action obvious.

- Date and warm greeting.
- Prominent “tonight’s plan” card with one primary action.
- At-a-glance inventory, use-soon, and preference metrics.
- Household constraints: peanut-free, dairy-light, high protein, and less food waste.
- After a confirmed meal, an in-session success banner names the meal and confirms inventory/history updates.

### 2. Choose a Meal (`view = choose`)

**Purpose:** provide a small, confidence-building choice set.

- Exactly three visually distinct fixtures: miso salmon bowls, creamy lemon chickpea pasta, and smoky chicken taco tray.
- Each option includes time, difficulty, description, recommendation rationale, goal tags, and a cook action.
- Safety reassurance confirms all options are peanut-free and goal-aligned.
- Back returns to Today without mutating data.

### 3. Cook (`view = cook`)

**Purpose:** reduce cognitive load while preparing and cooking.

- Recipe summary with time, difficulty, and serving count.
- Ingredient checklist shows required amounts and whether enough inventory exists. Checklist state is ephemeral.
- One current step receives visual emphasis; Previous/Next actions reveal sequential progress.
- The final step exposes Finish cooking. Back returns to meal choices without inventory changes.

### 4. Meal confirmation (`view = confirm`)

**Purpose:** close the loop transparently.

- Feedback is a required single choice: loved, okay, or not for us; default is loved.
- An optional note accepts up to 180 characters.
- The inventory preview displays every recipe ingredient, amount used, before amount, and deterministic after amount.
- Back returns to the intact recipe. Confirm commits both the history entry and inventory changes, clears the selected meal, and returns to Today.

### 5. Inventory (`view = inventory`)

**Purpose:** make the recommendation input inspectable.

- Current quantity, unit, and category for every stored ingredient.
- Search is case-insensitive. Category filters include All, Produce, Protein, Pantry, and Dairy.
- Items at or below their `lowAt` threshold show “Running low.” Empty filter results have an explicit state.

### 6. History (`view = history`)

**Purpose:** preserve household learning and demonstrate the feedback loop.

- Reverse-chronological meal cards show date, meal, rating, and optional note.
- Summary shows the number of loved meals.
- New confirmations appear immediately and survive reload.

## State and transitions

| From | Event | To | Persisted mutation |
|---|---|---|---|
| Today | Choose tonight’s meal | Choose | None |
| Choose | Cook this meal | Cook | `selectedMealId` |
| Cook | Finish cooking | Confirm | None |
| Confirm | Back to recipe | Cook | None |
| Confirm | Confirm meal | Today | Decrement inventory; prepend history; clear selection |
| Any | Main navigation | Today / Inventory / History | None |

If a persisted selected meal cannot be resolved, the app safely renders Today. View and checklist/current-step state are intentionally not persisted.

## Data model

The stored `AppState` uses the local-storage key `cyber-kitchen-demo-v1`.

```ts
type AppState = {
  inventory: InventoryItem[]
  history: HistoryEntry[]
  selectedMealId: string | null
}

type InventoryItem = {
  id: string
  name: string
  amount: number
  unit: string
  category: 'Produce' | 'Protein' | 'Pantry' | 'Dairy'
  lowAt: number
}

type HistoryEntry = {
  id: string
  mealId: string
  mealName: string
  emoji: string
  cookedAt: string // ISO timestamp
  rating: 'loved' | 'okay' | 'not-for-us'
  note: string
}
```

Recommendations are static `Meal` fixtures with identity, presentation fields, rationale, tags, duration, difficulty, ingredients, and ordered steps. A `MealIngredient` references inventory by `inventoryId` and supplies an amount/unit.

## Deterministic inventory rules

For every inventory item, confirmation finds the matching ingredient amount and computes `max(0, current amount - used amount)`. Unreferenced items are copied unchanged. The same `inventoryAfterMeal` pure function powers both preview and commit, preventing drift between what the person approves and what is stored.

## Persistence and demo behavior

- App state loads once from local storage and saves after state changes.
- Malformed stored JSON falls back to initial fixtures.
- No data leaves the browser and there is no authentication, API, telemetry, or live AI.
- Clearing the versioned local-storage key resets the demo.

## Responsive and accessible behavior

- Desktop uses a persistent sidebar; screens under 700px use a fixed bottom navigation bar.
- Intermediate layouts collapse choice cards and split panels before compact mobile layouts.
- Controls use semantic buttons, inputs, labels, radio roles, status regions, visible focus rings, and descriptive navigation labels.
- Color is reinforced by text and icons; `prefers-reduced-motion` disables transitions.
