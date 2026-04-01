import type { Combo } from '../types/combo'
import { useState, useRef, useEffect } from 'react'

interface ComboCardProps {
  combo: Combo
  onClick: () => void
  onRename: () => void
  onShare: () => void
  onDelete: () => void
}

export function ComboSummaryCard({ combo, onClick, onRename, onShare, onDelete }: ComboCardProps) {
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  // Close menu on outside click
  useEffect(() => {
    function handleOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }
    if (menuOpen) document.addEventListener('mousedown', handleOutside)
    return () => document.removeEventListener('mousedown', handleOutside)
  }, [menuOpen])

  const thumbnails = combo.items.slice(0, 3)
  const savedDate = new Date().toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })

  return (
    <div className="combo-card">
      {/* Thumbnail strip */}
      <div
        className="combo-card-thumbnails"
        onClick={onClick}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => e.key === 'Enter' && onClick()}
      >
        {thumbnails.length > 0
          ? thumbnails.map((item) => (
              <img key={item.simpleSku} src={item.imageUrl} alt={item.name} />
            ))
          : <div className="thumb-placeholder">🖼️</div>}
      </div>

      {/* Card body */}
      <div className="combo-card-body" onClick={onClick} style={{ cursor: 'pointer' }}>
        <div className="combo-card-name">{combo.name}</div>
        <div className="combo-card-meta">{combo.items.length} items · Saved {savedDate}</div>
      </div>

      {/* Footer */}
      <div className="combo-card-footer">
        <span className={`combo-card-badge ${combo.visibility === 'public' ? 'public' : ''}`}>
          {combo.visibility === 'public' ? '🔗 Public' : '🔒 Private'}
        </span>

        <div className="menu-wrapper" ref={menuRef}>
          <button
            className="menu-btn"
            aria-label="Options"
            onClick={(e) => { e.stopPropagation(); setMenuOpen((o) => !o) }}
          >
            ⋯
          </button>
          {menuOpen && (
            <div className="dropdown-menu" role="menu">
              <button className="dropdown-item" onClick={() => { setMenuOpen(false); onRename() }}>
                ✏️ Rename
              </button>
              <button className="dropdown-item" onClick={() => { setMenuOpen(false); onShare() }}>
                🔗 Share
              </button>
              <button className="dropdown-item danger" onClick={() => { setMenuOpen(false); onDelete() }}>
                🗑 Delete
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
