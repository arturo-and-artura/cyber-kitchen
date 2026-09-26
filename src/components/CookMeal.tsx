import { ArrowLeft, ArrowRight, Check, CheckCircle2, ChefHat, Clock3, UtensilsCrossed } from 'lucide-react'
import { useState } from 'react'
import type { InventoryItem, Meal } from '../types'
import { formatAmount } from '../lib/store'

type Props = { meal: Meal; inventory: InventoryItem[]; onBack: () => void; onFinish: () => void }

export function CookMeal({ meal, inventory, onBack, onFinish }: Props) {
  const [checked, setChecked] = useState<string[]>([])
  const [step, setStep] = useState(0)
  const toggle = (id: string) => setChecked((current) => current.includes(id) ? current.filter((x) => x !== id) : [...current, id])
  return (
    <div className="page cook-page">
      <button className="back-button" onClick={onBack}><ArrowLeft size={17} /> Meal choices</button>
      <header className="cook-header"><div className={`mini-art ${meal.accent}`}><UtensilsCrossed /></div><div><span className="eyebrow">COOKING MODE</span><h1>{meal.name}</h1><p><Clock3 size={15} /> {meal.minutes} minutes · {meal.difficulty} · Serves 3</p></div></header>
      <div className="cook-layout">
        <section className="panel ingredient-panel"><div className="panel-heading"><div><span className="eyebrow">GET READY</span><h2>Ingredients</h2></div><span>{checked.length}/{meal.ingredients.length} ready</span></div>
          <div className="ingredient-list">
            {meal.ingredients.map((ingredient) => {
              const item = inventory.find((i) => i.id === ingredient.inventoryId)
              const available = item && item.amount >= ingredient.amount
              return <label key={ingredient.inventoryId} className={checked.includes(ingredient.inventoryId) ? 'checked' : ''}>
                <input type="checkbox" checked={checked.includes(ingredient.inventoryId)} onChange={() => toggle(ingredient.inventoryId)} />
                <span className="custom-check"><Check size={14} /></span><span><strong>{ingredient.name}</strong><small>{formatAmount(ingredient.amount, ingredient.unit)}</small></span><b className={available ? 'available' : 'missing'}>{available ? 'On hand' : 'Need'}</b>
              </label>
            })}
          </div>
        </section>
        <section className="panel steps-panel"><div className="panel-heading"><div><span className="eyebrow">STEP BY STEP</span><h2>Let’s cook</h2></div><span>{step + 1} of {meal.steps.length}</span></div>
          <div className="progress-track"><span style={{ width: `${((step + 1) / meal.steps.length) * 100}%` }} /></div>
          <div className="current-step"><span>{step + 1}</span><p>{meal.steps[step]}</p></div>
          <ol className="step-list">{meal.steps.map((text, index) => <li key={text} className={index < step ? 'done' : index === step ? 'current' : ''}><span>{index < step ? <Check size={14} /> : index + 1}</span><p>{text}</p></li>)}</ol>
          <div className="step-actions">
            <button className="secondary-button" disabled={step === 0} onClick={() => setStep((s) => s - 1)}>Previous</button>
            {step < meal.steps.length - 1 ? <button className="primary-button" onClick={() => setStep((s) => s + 1)}>Next step <ArrowRight size={17} /></button> : <button className="primary-button" onClick={onFinish}><ChefHat size={17} /> Finish cooking</button>}
          </div>
          {step === meal.steps.length - 1 && <p className="finish-hint"><CheckCircle2 size={16} /> You made it — serve when ready!</p>}
        </section>
      </div>
    </div>
  )
}
