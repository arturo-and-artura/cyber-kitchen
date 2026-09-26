import { ArrowRight, Check, Clock3, Leaf, PackageOpen, Sparkles, UtensilsCrossed, Users } from 'lucide-react'
import type { InventoryItem } from '../types'

type Props = { inventory: InventoryItem[]; cookedTonight?: string; onChoose: () => void }

export function Today({ inventory, cookedTonight, onChoose }: Props) {
  const lowItems = inventory.filter((item) => item.amount <= item.lowAt).length
  return (
    <div className="page today-page">
      <header className="page-header">
        <div><p className="date-label">SATURDAY, SEPTEMBER 26</p><h1>Good morning, kitchen crew.</h1><p>Let’s make tonight feel easy.</p></div>
        <div className="status-pill"><span className="pulse" /> Inventory synced</div>
      </header>

      {cookedTonight && (
        <div className="success-banner" role="status"><span><Check size={18} /></span><div><strong>Dinner is done!</strong><p>{cookedTonight} was added to your history and your inventory is up to date.</p></div></div>
      )}

      <section className="hero-card">
        <div className="hero-copy">
          <span className="eyebrow light"><Sparkles size={15} /> TONIGHT'S PLAN</span>
          <h2>What sounds good<br />for dinner?</h2>
          <p>We found three ideas based on what you have, what your household enjoys, and your goals.</p>
          <button className="primary-button inverted" onClick={onChoose}>Choose tonight’s meal <ArrowRight size={18} /></button>
          <span className="helper"><Clock3 size={15} /> All options take 35 minutes or less</span>
        </div>
        <div className="hero-art" aria-hidden="true">
          <span className="plate"><UtensilsCrossed /></span><span className="floating tomato"><Sparkles /></span><span className="floating leaf"><Leaf /></span><span className="floating lemon"><Sparkles /></span>
        </div>
      </section>

      <div className="section-heading"><div><span className="eyebrow">AT A GLANCE</span><h2>Your kitchen today</h2></div></div>
      <div className="stats-grid">
        <article className="stat-card"><span className="stat-icon green"><PackageOpen /></span><div><strong>{inventory.length}</strong><p>items on hand</p></div><small>{lowItems ? `${lowItems} running low` : 'All well stocked'}</small></article>
        <article className="stat-card"><span className="stat-icon coral"><Clock3 /></span><div><strong>3</strong><p>items to use soon</p></div><small>We’ll prioritize these</small></article>
        <article className="stat-card"><span className="stat-icon yellow"><Users /></span><div><strong>4</strong><p>shared preferences</p></div><small>Applied to every idea</small></article>
      </div>
      <section className="constraint-strip"><span>Safety & goals always on</span><div><b>✓ Peanut-free</b><b>◌ Dairy-light</b><b>↗ High protein</b><b>♻ Less food waste</b></div></section>
    </div>
  )
}
