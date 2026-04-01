import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { comboApi } from '../api/comboApi'
import type { Combo } from '../types/combo'
import { ComboSummaryCard } from '../components/ComboSummaryCard'
import { RenameModal } from '../components/RenameModal'
import { ShareSheet } from '../components/ShareSheet'
import { useToast } from '../components/Toast'

export function MyCombosPage() {
  const navigate = useNavigate()
  const [combos, setCombos] = useState<Combo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [renaming, setRenaming] = useState<Combo | null>(null)
  const [sharing, setSharing] = useState<{ combo: Combo; shareUrl: string } | null>(null)
  const { showToast } = useToast()

  const loadCombos = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await comboApi.list()
      setCombos(data ?? [])
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadCombos() }, [loadCombos])

  async function handleShare(combo: Combo) {
    try {
      const result = await comboApi.share(combo.id)
      setSharing({ combo, shareUrl: result.shareUrl })
    } catch (e) {
      showToast('Failed to generate share link')
    }
  }

  async function handleDelete(combo: Combo) {
    if (!confirm(`Delete "${combo.name}"? This cannot be undone.`)) return
    try {
      await comboApi.remove(combo.id)
      setCombos((prev) => prev.filter((c) => c.id !== combo.id))
      showToast('Combo deleted')
    } catch (e) {
      showToast('Failed to delete combo')
    }
  }

  async function handleRenameConfirm(newName: string) {
    if (!renaming) return
    try {
      await comboApi.rename(renaming.id, newName)
      setCombos((prev) => prev.map((c) => c.id === renaming.id ? { ...c, name: newName } : c))
      showToast('Combo renamed')
    } catch (e) {
      showToast('Failed to rename combo')
    } finally {
      setRenaming(null)
    }
  }

  return (
    <div>
      <header className="page-header">
        <h1>My Combos {combos.length > 0 && <span style={{ fontWeight: 400, color: 'var(--color-text-secondary)' }}>({combos.length})</span>}</h1>
      </header>

      <main className="page-container">
        {loading && <div className="spinner" />}

        {error && <div className="error-banner">⚠️ {error}</div>}

        {!loading && !error && combos.length === 0 && (
          <div className="empty-state">
            <div className="empty-icon">🛍️</div>
            <h2>No combos saved yet</h2>
            <p>Generate your first combo from your wishlist and save it here.</p>
          </div>
        )}

        {!loading && combos.length > 0 && (
          <div className="combos-grid">
            {combos.map((combo) => (
              <ComboSummaryCard
                key={combo.id}
                combo={combo}
                onClick={() => navigate(`/combos/${combo.id}`)}
                onRename={() => setRenaming(combo)}
                onShare={() => handleShare(combo)}
                onDelete={() => handleDelete(combo)}
              />
            ))}
          </div>
        )}
      </main>

      {renaming && (
        <RenameModal
          currentName={renaming.name}
          onConfirm={handleRenameConfirm}
          onClose={() => setRenaming(null)}
        />
      )}

      {sharing && (
        <ShareSheet
          shareUrl={sharing.shareUrl}
          comboName={sharing.combo.name}
          onClose={() => setSharing(null)}
        />
      )}
    </div>
  )
}
