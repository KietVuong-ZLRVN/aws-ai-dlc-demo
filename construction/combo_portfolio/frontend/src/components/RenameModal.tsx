import { useState, useRef, useEffect } from 'react'

interface RenameModalProps {
  currentName: string
  onConfirm: (name: string) => void
  onClose: () => void
}

export function RenameModal({ currentName, onConfirm, onClose }: RenameModalProps) {
  const [name, setName] = useState(currentName)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    inputRef.current?.select()
  }, [])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (name.trim().length < 2) return
    onConfirm(name.trim())
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2 className="modal-title">Rename combo</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label" htmlFor="combo-name">Combo name</label>
            <input
              id="combo-name"
              ref={inputRef}
              className="form-input"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={100}
              autoFocus
            />
          </div>
          <div className="modal-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose}>Cancel</button>
            <button type="submit" className="btn btn-primary" disabled={name.trim().length < 2}>
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
