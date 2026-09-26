import { describe, expect, it } from 'vitest'
import { initialState, meals } from '../data/mockData'
import { formatAmount, inventoryAfterMeal } from './store'

describe('inventoryAfterMeal', () => {
  it('subtracts recipe amounts without mutating the original inventory', () => {
    const result = inventoryAfterMeal(initialState, meals[0])
    expect(result.find((item) => item.id === 'salmon')?.amount).toBe(0)
    expect(result.find((item) => item.id === 'rice')?.amount).toBe(2)
    expect(initialState.inventory.find((item) => item.id === 'salmon')?.amount).toBe(2)
  })

  it('never produces a negative amount', () => {
    const depleted = { ...initialState, inventory: initialState.inventory.map((item) => ({ ...item, amount: 0 })) }
    expect(inventoryAfterMeal(depleted, meals[0]).every((item) => item.amount >= 0)).toBe(true)
  })
})

describe('formatAmount', () => {
  it('renders integers and decimals with their units', () => {
    expect(formatAmount(2, 'cups')).toBe('2 cups')
    expect(formatAmount(1.5, 'tbsp')).toBe('1.5 tbsp')
  })
})
