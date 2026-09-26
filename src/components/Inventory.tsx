import { AlertCircle, PackageOpen, Search } from 'lucide-react'
import { useMemo, useState } from 'react'
import type { InventoryItem } from '../types'
import { formatAmount } from '../lib/store'

export function Inventory({ inventory }: { inventory: InventoryItem[] }) {
  const [query, setQuery] = useState('')
  const [category, setCategory] = useState('All')
  const visible = useMemo(() => inventory.filter((item) => (category === 'All' || item.category === category) && item.name.toLowerCase().includes(query.toLowerCase())), [inventory, query, category])
  const categories = ['All', 'Produce', 'Protein', 'Pantry', 'Dairy']
  return <div className="page">
    <header className="page-header"><div><p className="date-label">YOUR KITCHEN</p><h1>Inventory</h1><p>A simple, shared picture of what’s on hand.</p></div><div className="status-pill"><span className="pulse" /> Demo inventory</div></header>
    <div className="toolbar"><label className="search"><Search size={18} /><span className="sr-only">Search inventory</span><input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search ingredients" /></label><div className="filter-tabs">{categories.map((item) => <button className={category === item ? 'active' : ''} onClick={() => setCategory(item)} key={item}>{item}</button>)}</div></div>
    <div className="inventory-grid">{visible.map((item) => <article key={item.id} className="inventory-card"><span className={`category-icon ${item.category.toLowerCase()}`}><PackageOpen size={20} /></span><div><h2>{item.name}</h2><p>{item.category}</p></div><strong>{formatAmount(item.amount, item.unit)}</strong>{item.amount <= item.lowAt && <span className="low-badge"><AlertCircle size={13} /> Running low</span>}</article>)}</div>
    {!visible.length && <div className="empty-state"><PackageOpen /><h2>No ingredients found</h2><p>Try another search or category.</p></div>}
  </div>
}
