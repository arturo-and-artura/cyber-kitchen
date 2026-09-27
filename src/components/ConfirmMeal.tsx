import { Check, Heart, Meh, RotateCcw, Sparkles, ThumbsDown } from 'lucide-react'
import { useState } from 'react'
import type { AppState, Meal } from '../types'
import { formatAmount, inventoryAfterMeal } from '../lib/store'

type Rating = 'loved' | 'okay' | 'not-for-us'
type Props = { meal: Meal; state: AppState; onConfirm: (rating: Rating, note: string) => void; onBack: () => void }

export function ConfirmMeal({ meal, state, onConfirm, onBack }: Props) {
  const [rating, setRating] = useState<Rating>('loved')
  const [note, setNote] = useState('')
  const after = inventoryAfterMeal(state, meal)
  return (
    <div className="page confirm-page">
      <header className="celebrate"><div className={`celebrate-art ${meal.accent}`}><Sparkles /></div><span className="eyebrow">DINNER COMPLETE</span><h1>Nice work, kitchen crew!</h1><p>Before we tidy up, tell us how it went and review what will change.</p></header>
      <div className="confirm-layout">
        <section className="panel feedback-panel"><span className="eyebrow">QUICK FEEDBACK</span><h2>How was {meal.name}?</h2>
          <div className="rating-grid" role="radiogroup" aria-label="Meal rating">
            {[{ id: 'loved', label: 'Loved it', Icon: Heart }, { id: 'okay', label: 'It was okay', Icon: Meh }, { id: 'not-for-us', label: 'Not for us', Icon: ThumbsDown }].map(({ id, label, Icon }) => <button role="radio" aria-checked={rating === id} key={id} className={rating === id ? 'selected' : ''} onClick={() => setRating(id as Rating)}><Icon size={23} /><span>{label}</span></button>)}
          </div>
          <label className="note-field">Anything to remember for next time? <span>Optional</span><textarea value={note} onChange={(event) => setNote(event.target.value)} placeholder="e.g. Everyone loved the extra cucumber…" maxLength={180} /></label>
        </section>
        <section className="panel adjustment-panel"><div className="panel-heading"><div><span className="eyebrow">INVENTORY PREVIEW</span><h2>What we’ll update</h2></div><span className="automatic"><RotateCcw size={13} /> Automatic</span></div>
          <p className="muted">Based on the recipe amounts. Nothing changes until you confirm.</p>
          <div className="adjustment-list">{meal.ingredients.map((ingredient) => {
            const before = state.inventory.find((item) => item.id === ingredient.inventoryId)
            const next = after.find((item) => item.id === ingredient.inventoryId)
            return <div key={ingredient.inventoryId}><span><strong>{ingredient.name}</strong><small>Used {formatAmount(ingredient.amount, ingredient.unit)}</small></span><span className="amount-change"><s>{before ? formatAmount(before.amount, before.unit) : '0'}</s><b>→</b><strong>{next ? formatAmount(next.amount, next.unit) : '0'}</strong></span></div>
          })}</div>
        </section>
      </div>
      <div className="confirm-actions"><button className="secondary-button" onClick={onBack}>Back to recipe</button><button className="primary-button large" onClick={() => onConfirm(rating, note)}><Check size={18} /> Confirm meal & update inventory</button></div>
    </div>
  )
}
