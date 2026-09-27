import { useEffect, useMemo, useState } from 'react'
import { ChooseMeal } from './components/ChooseMeal'
import { ConfirmMeal } from './components/ConfirmMeal'
import { CookMeal } from './components/CookMeal'
import { History } from './components/History'
import { Inventory } from './components/Inventory'
import { Shell } from './components/Shell'
import { Today } from './components/Today'
import { meals } from './data/mockData'
import { inventoryAfterMeal, loadState, STORAGE_KEY } from './lib/store'
import type { AppState, Meal, View } from './types'

export default function App() {
  const [view, setView] = useState<View>('today')
  const [state, setState] = useState<AppState>(loadState)
  const [latestMeal, setLatestMeal] = useState<string>()
  const selectedMeal = useMemo(() => meals.find((meal) => meal.id === state.selectedMealId) ?? null, [state.selectedMealId])

  useEffect(() => localStorage.setItem(STORAGE_KEY, JSON.stringify(state)), [state])

  const selectMeal = (meal: Meal) => { setState((current) => ({ ...current, selectedMealId: meal.id })); setView('cook'); window.scrollTo(0, 0) }
  const confirm = (rating: 'loved' | 'okay' | 'not-for-us', note: string) => {
    if (!selectedMeal) return
    setState((current) => ({
      inventory: inventoryAfterMeal(current, selectedMeal),
      selectedMealId: null,
      history: [{ id: `history-${Date.now()}`, mealId: selectedMeal.id, mealName: selectedMeal.name, emoji: selectedMeal.emoji, cookedAt: new Date().toISOString(), rating, note }, ...current.history],
    }))
    setLatestMeal(selectedMeal.name); setView('today'); window.scrollTo(0, 0)
  }

  let content
  if (view === 'choose') content = <ChooseMeal meals={meals} onBack={() => setView('today')} onSelect={selectMeal} />
  else if (view === 'cook' && selectedMeal) content = <CookMeal meal={selectedMeal} inventory={state.inventory} onBack={() => setView('choose')} onFinish={() => setView('confirm')} />
  else if (view === 'confirm' && selectedMeal) content = <ConfirmMeal meal={selectedMeal} state={state} onBack={() => setView('cook')} onConfirm={confirm} />
  else if (view === 'inventory') content = <Inventory inventory={state.inventory} />
  else if (view === 'history') content = <History history={state.history} />
  else content = <Today inventory={state.inventory} cookedTonight={latestMeal} onChoose={() => setView('choose')} />

  return <Shell view={view} setView={setView}>{content}</Shell>
}
