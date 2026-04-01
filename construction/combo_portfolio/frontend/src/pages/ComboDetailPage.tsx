import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { comboApi } from '../api/comboApi'
import type { Combo } from '../types/combo'
import { RenameModal } from '../components/RenameModal'
import { ShareSheet } from '../components/ShareSheet'
import { useToast } from '../components/Toast'

export function ComboDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [combo, setCombo] = useState<Combo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [menuOpen, setMenuOpen] = useState(false)
  const [renaming, setRenaming] = useState(false)
  const [shareUrl, setShareUrl] = useState<string | null>(null)
  const [addingToCart, setAddingToCart] = useState(false)
  const { showToast } = useToast()

  const loadCombo = useCallback(async () => {
    if (!id) return
    setLoading(true)
    setError(null)
    try {
      const data = await comboApi.get(id)
      setCombo(data)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => { loadCombo() }, [loadCombo])

  async function handleShare() {
    if (!combo) return
    setMenuOpen(false)
    try {
      const result = await comboApi.share(combo.id)
      setShareUrl(result.shareUrl)
      setCombo((c) => c ? { ...c, visibility: 'public', shareToken: result.shareToken } : c)
    } catch (e) {
      showToast('Failed to generate share link')
    }
  }

  async function handleVisibilityToggle() {
    if (!combo) return
    const next = combo.visibility === 'public' ? 'private' : 'public'
    // To make public, must share first
    if (next === 'public') { handleShare(); return }
    try {
      await comboApi.setVisibility(combo.id, 'private')
      setCombo((c) => c ? { ...c, visibility: 'private', shareToken: undefined } : c)
      showToast('Combo set to private')
    } catch (e) {
      showToast('Failed to update visibility')
    }
  }

  async function handleRename(newName: string) {
    if (!combo) return
    try {
      await comboApi.rename(combo.id, newName)
      setCombo((c) => c ? { ...c, name: newName } : c)
      showToast('Combo renamed')
    } catch (e) {
      showToast('Failed to rename combo')
    } finally {
      setRenaming(false)
    }
  }

  async function handleDelete() {
    if (!combo) return
    setMenuOpen(false)
    if (!confirm(`Delete "${combo.name}"? This cannot be undone.`)) return
    try {
      await comboApi.remove(combo.id)
      showToast('Combo deleted')
      navigate('/')
    } catch (e) {
      showToast('Failed to delete combo')
    }
  }

  async function handleAddAllToCart() {
    if (!combo) return
    setAddingToCart(true)
    try {
      // Calls Unit 5: POST /api/v1/cart/combo
      const res = await fetch('/api/v1/cart/combo', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ comboId: combo.id }),
      })
      const body = await res.json()
      if (!res.ok) {
        showToast(body.error ?? 'Failed to add to cart')
        return
      }
      const status: string = body.status
      if (status === 'ok') {
        showToast(`All ${combo.items.length} items added to cart ✓`)
      } else if (status === 'partial') {
        const skipped = body.skippedItems?.length ?? 0
        showToast(`Added ${body.addedItems?.length ?? 0} items — ${skipped} out of stock`)
      } else {
        showToast('Could not add items to cart')
      }
    } catch {
      showToast('Cart service unavailable')
    } finally {
      setAddingToCart(false)
    }
  }

  if (loading) return <div className="spinner" />
  if (error) return (
    <div className="page-container">
      <div className="error-banner">⚠️ {error}</div>
    </div>
  )
  if (!combo) return null

  return (
    <div>
      {/* Sticky header */}
      <header className="detail-header">
        <button className="back-btn" onClick={() => navigate('/')} aria-label="Back">
          ←
        </button>
        <span className="detail-name">{combo.name}</span>
        <div className="detail-actions">
          <button className="btn btn-ghost" onClick={handleShare}>🔗 Share</button>
          <div style={{ position: 'relative' }}>
            <button
              className="menu-btn"
              aria-label="Options"
              style={{ padding: '8px', borderRadius: '50%', border: 'none', cursor: 'pointer', background: 'none' }}
              onClick={() => setMenuOpen((o) => !o)}
            >
              ⋯
            </button>
            {menuOpen && (
              <div className="dropdown-menu" role="menu" style={{ right: 0, top: '40px' }}>
                <button className="dropdown-item" onClick={() => { setMenuOpen(false); setRenaming(true) }}>✏️ Rename</button>
                <button className="dropdown-item danger" onClick={handleDelete}>🗑 Delete</button>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* Item list */}
      <div className="item-list">
        {combo.items.map((item) => (
          <div key={item.simpleSku} className="item-row">
            <img src={item.imageUrl} alt={item.name} className="item-image" />
            <div className="item-info">
              <div className="item-name">{item.name}</div>
              <div className="item-price">S${item.price.toFixed(2)}</div>
              <div className="item-sku">{item.simpleSku}</div>
              <span className={`item-badge ${item.inStock ? 'in-stock' : ''}`}>
                {item.inStock ? '● In stock' : '○ Out of stock'}
              </span>
            </div>
          </div>
        ))}
      </div>

      {/* Sticky footer */}
      <div className="detail-footer">
        {/* Visibility toggle */}
        <div className="visibility-toggle">
          <label className="toggle">
            <input
              type="checkbox"
              checked={combo.visibility === 'public'}
              onChange={handleVisibilityToggle}
            />
            <span className="toggle-slider" />
          </label>
          <span className="visibility-label">
            {combo.visibility === 'public' ? 'Public (shareable)' : 'Private'}
          </span>
        </div>

        {/* Add All to Cart */}
        <button
          className="btn btn-primary btn-full btn-lg"
          onClick={handleAddAllToCart}
          disabled={addingToCart}
        >
          {addingToCart ? 'Adding to cart…' : `Add All ${combo.items.length} Items to Cart`}
        </button>
      </div>

      {renaming && (
        <RenameModal
          currentName={combo.name}
          onConfirm={handleRename}
          onClose={() => setRenaming(false)}
        />
      )}

      {shareUrl && (
        <ShareSheet
          shareUrl={shareUrl}
          comboName={combo.name}
          onClose={() => setShareUrl(null)}
        />
      )}
    </div>
  )
}
