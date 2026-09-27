import { ArrowLeft, ArrowRight, Check, Clock3, Sparkles } from 'lucide-react'
import type { Meal } from '../types'

type Props = { meals: Meal[]; onBack: () => void; onSelect: (meal: Meal) => void }

export function ChooseMeal({ meals, onBack, onSelect }: Props) {
  return (
    <div className="page choose-page">
      <button className="back-button" onClick={onBack}><ArrowLeft size={17} /> Today</button>
      <header className="choice-header"><span className="eyebrow"><Sparkles size={14} /> THREE THOUGHTFUL MATCHES</span><h1>Pick tonight’s dinner</h1><p>Every option respects your household profile. We’ve explained why each one made the list.</p></header>
      <div className="meal-grid">
        {meals.map((meal, index) => (
          <article className="meal-card" key={meal.id}>
            <div className={`meal-art ${meal.accent}`}><span>{meal.name.split(' ')[0]}</span>{index === 0 && <b>Best match</b>}</div>
            <div className="meal-body">
              <div className="meal-meta"><span><Clock3 size={14} /> {meal.minutes} min</span><span>{meal.difficulty}</span></div>
              <h2>{meal.name}</h2><p>{meal.description}</p>
              <div className="why"><Sparkles size={17} /><div><strong>Why this works tonight</strong><p>{meal.reason}</p></div></div>
              <div className="tag-row">{meal.tags.map((tag) => <span key={tag}><Check size={13} />{tag}</span>)}</div>
              <button className="primary-button" onClick={() => onSelect(meal)}>Cook this meal <ArrowRight size={17} /></button>
            </div>
          </article>
        ))}
      </div>
      <p className="safety-note"><Check size={15} /> All three meals are peanut-free and aligned with your household goals.</p>
    </div>
  )
}
