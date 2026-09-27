import { initialState } from '../data/mockData'
import type { AppState, Meal } from '../types'

export const STORAGE_KEY = 'cyber-kitchen-demo-v1'

export function loadState(): AppState {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    return saved ? (JSON.parse(saved) as AppState) : initialState
  } catch {
    return initialState
  }
}

export function inventoryAfterMeal(state: AppState, meal: Meal) {
  return state.inventory.map((item) => {
    const used = meal.ingredients.find((ingredient) => ingredient.inventoryId === item.id)?.amount ?? 0
    return { ...item, amount: Math.max(0, item.amount - used) }
  })
}

export function formatAmount(amount: number, unit: string) {
  return `${Number.isInteger(amount) ? amount : amount.toFixed(1)}${unit ? ` ${unit}` : ''}`
}
