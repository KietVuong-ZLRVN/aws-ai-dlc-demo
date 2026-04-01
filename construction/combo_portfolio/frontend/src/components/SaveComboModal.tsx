import { useState } from 'react'

interface SaveComboModalProps {
  /** Pre-built item list from the AI engine (Unit 3) */
  items: Array<{
    configSku: string
    simpleSku: string
    name: string
    imageUrl: string
    price: number
  }>
  onSave: (name: string) => Promise<void>
  onClose: () => void
}

function defaultName() {
  return `Combo – ${new Date().toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })}`
}

export function SaveComboModal({ items, onSave, onClose }: SaveComboModalProps) {
  const [name, setName] = useState(defaultName())
  const [saving, setSaving] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (name.trim().length < 2) return
    setSaving(true)
    try {
      await onSave(name.trim())
      onClose()
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2 className="modal-title">Save this combo</h2>

        {/* Item preview strip */}
        <div style={{ display: 'flex', gap: 8, marginBottom: 20 }}>
          {items.slice(0, 3).map((item) => (
            <img
              key={item.simpleSku}
              src={item.imageUrl}
              alt={item.name}
              style={{ width: 64, height: 64, objectFit: 'cover', borderRadius: 8, background: '#e5e5e0' }}
            />
          ))}
          {items.length > 3 && (
            <div style={{
              width: 64, height: 64, borderRadius: 8, background: '#e5e5e0',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 13, color: '#6b6b6b', fontWeight: 600,
            }}>
              +{items.length - 3}
            </div>
          )}
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label" htmlFor="save-name">Combo name</label>
            <input
              id="save-name"
              className="form-input"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={100}
              autoFocus
            />
          </div>

          <div className="modal-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={saving}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={saving || name.trim().length < 2}>
              {saving ? 'Saving…' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
