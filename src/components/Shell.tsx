import { CalendarDays, ChefHat, Clock3, PackageOpen, Sparkles } from 'lucide-react'
import type { ReactNode } from 'react'
import type { View } from '../types'

type Props = { view: View; setView: (view: View) => void; children: ReactNode }

export function Shell({ view, setView, children }: Props) {
  const activeView = view === 'choose' || view === 'cook' || view === 'confirm' ? 'today' : view
  const nav = [
    { id: 'today' as const, label: 'Today', icon: CalendarDays },
    { id: 'inventory' as const, label: 'Inventory', icon: PackageOpen },
    { id: 'history' as const, label: 'History', icon: Clock3 },
  ]

  return (
    <div className="app-shell">
      <aside className="sidebar" aria-label="Main navigation">
        <button className="brand" onClick={() => setView('today')} aria-label="Cyber Kitchen home">
          <span className="brand-mark"><ChefHat size={22} /></span>
          <span><strong>Cyber Kitchen</strong><small>Thoughtful meals, together</small></span>
        </button>
        <nav>
          {nav.map(({ id, label, icon: Icon }) => (
            <button key={id} className={activeView === id ? 'nav-item active' : 'nav-item'} onClick={() => setView(id)}>
              <Icon size={19} /> {label}
            </button>
          ))}
        </nav>
        <div className="household-card">
          <span className="eyebrow"><Sparkles size={13} /> Household profile</span>
          <strong>The Lee household</strong>
          <div className="avatar-row" aria-label="Household members"><span>YL</span><span>AL</span><span>ML</span></div>
          <p><b>Peanut-free</b> · Dairy-light<br />High protein · Less food waste</p>
        </div>
      </aside>
      <main id="main-content">{children}</main>
      <nav className="mobile-nav" aria-label="Main navigation">
        {nav.map(({ id, label, icon: Icon }) => (
          <button key={id} className={activeView === id ? 'active' : ''} onClick={() => setView(id)}>
            <Icon size={20} /><span>{label}</span>
          </button>
        ))}
      </nav>
    </div>
  )
}
