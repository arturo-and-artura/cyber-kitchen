export type View = 'today' | 'choose' | 'cook' | 'confirm' | 'inventory' | 'history'

export type InventoryItem = {
  id: string
  name: string
  amount: number
  unit: string
  category: 'Produce' | 'Protein' | 'Pantry' | 'Dairy'
  lowAt: number
}

export type MealIngredient = {
  inventoryId: string
  name: string
  amount: number
  unit: string
  optional?: boolean
}

export type Meal = {
  id: string
  name: string
  description: string
  reason: string
  emoji: string
  accent: string
  minutes: number
  difficulty: 'Easy' | 'Medium'
  tags: string[]
  ingredients: MealIngredient[]
  steps: string[]
}

export type HistoryEntry = {
  id: string
  mealId: string
  mealName: string
  emoji: string
  cookedAt: string
  rating: 'loved' | 'okay' | 'not-for-us'
  note: string
}

export type AppState = {
  inventory: InventoryItem[]
  history: HistoryEntry[]
  selectedMealId: string | null
}
