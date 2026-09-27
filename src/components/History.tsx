import { CalendarDays, Heart, Meh, Sparkles, ThumbsDown, UtensilsCrossed } from 'lucide-react'
import type { HistoryEntry } from '../types'

const icons = { loved: Heart, okay: Meh, 'not-for-us': ThumbsDown }
const labels = { loved: 'Loved it', okay: 'It was okay', 'not-for-us': 'Not for us' }

export function History({ history }: { history: HistoryEntry[] }) {
  return <div className="page">
    <header className="page-header"><div><p className="date-label">YOUR TABLE</p><h1>Meal history</h1><p>Little notes that make the next recommendation better.</p></div></header>
    <section className="history-summary"><Sparkles size={22} /><div><strong>{history.filter((h) => h.rating === 'loved').length} household favorites</strong><p>This demo uses your feedback to explain how future choices could improve.</p></div></section>
    <div className="timeline">{history.map((entry) => { const Icon = icons[entry.rating]; return <article key={entry.id} className="history-card"><div className="history-emoji"><UtensilsCrossed /></div><div className="history-content"><div><span><CalendarDays size={14} /> {new Intl.DateTimeFormat('en', { month: 'long', day: 'numeric', year: 'numeric' }).format(new Date(entry.cookedAt))}</span><span className={`rating ${entry.rating}`}><Icon size={14} /> {labels[entry.rating]}</span></div><h2>{entry.mealName}</h2>{entry.note && <blockquote>“{entry.note}”</blockquote>}</div></article> })}</div>
  </div>
}
