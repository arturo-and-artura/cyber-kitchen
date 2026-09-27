import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it } from 'vitest'
import App from './App'
import { STORAGE_KEY } from './lib/store'

describe('Cyber Kitchen primary flow', () => {
  beforeEach(() => localStorage.clear())

  it('chooses, cooks, confirms, and reflects a meal', async () => {
    const user = userEvent.setup()
    render(<App />)

    expect(screen.getByRole('heading', { name: /good morning/i })).toBeInTheDocument()
    expect(screen.getByText('✓ Peanut-free')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: /choose tonight’s meal/i }))

    expect(screen.getByRole('heading', { name: /pick tonight’s dinner/i })).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: /cook this meal/i })).toHaveLength(3)
    await user.click(screen.getAllByRole('button', { name: /cook this meal/i })[0])

    expect(screen.getByRole('heading', { name: 'Ingredients' })).toBeInTheDocument()
    for (let index = 0; index < 4; index += 1) await user.click(screen.getByRole('button', { name: /next step/i }))
    await user.click(screen.getByRole('button', { name: /finish cooking/i }))

    expect(screen.getByRole('heading', { name: /nice work/i })).toBeInTheDocument()
    expect(screen.getByText(/nothing changes until you confirm/i)).toBeInTheDocument()
    await user.type(screen.getByPlaceholderText(/everyone loved/i), 'Bright and easy.')
    await user.click(screen.getByRole('button', { name: /confirm meal & update inventory/i }))

    expect(screen.getByText('Dinner is done!')).toBeInTheDocument()
    expect(screen.getByText(/miso-glazed salmon bowls was added/i)).toBeInTheDocument()
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '{}')
    expect(saved.inventory.find((item: { id: string }) => item.id === 'salmon').amount).toBe(0)
    expect(saved.history[0].note).toBe('Bright and easy.')
  })

  it('filters inventory and shows history', async () => {
    const user = userEvent.setup()
    render(<App />)
    const nav = screen.getByRole('navigation', { name: 'Main navigation' })
    await user.click(within(nav).getByRole('button', { name: 'Inventory' }))
    await user.type(screen.getByPlaceholderText('Search ingredients'), 'salmon')
    expect(screen.getByRole('heading', { name: 'Salmon fillets' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Brown rice' })).not.toBeInTheDocument()
    await user.click(within(nav).getByRole('button', { name: 'History' }))
    expect(screen.getByRole('heading', { name: 'Meal history' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Roasted tomato soup' })).toBeInTheDocument()
  })
})
