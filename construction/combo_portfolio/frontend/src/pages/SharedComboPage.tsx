import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { comboApi } from '../api/comboApi'
import type { Combo } from '../types/combo'
import { ShareSheet } from '../components/ShareSheet'

export function SharedComboPage() {
  const { token } = useParams<{ token: string }>()
  const [combo, setCombo] = useState<Combo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showShare, setShowShare] = useState(false)
  const [addingToCart, setAddingToCart] = useState(false)
  const [cartStatus, setCartStatus] = useState<string | null>(null)

  useEffect(() => {
    if (!token) return
    comboApi.getShared(token)
      .then(setCombo)
      .catch((e) => setError((e as Error).message))
      .finally(() => setLoading(false))
  }, [token])

  async function handleAddAllToCart() {
    if (!combo) return
    setAddingToCart(true)
    setCartStatus(null)
    try {
      const res = await fetch('/api/v1/cart/combo', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          items: combo.items.map((i) => ({ simpleSku: i.simpleSku, quantity: 1 })),
        }),
      })
      const body = await res.json()
      if (!res.ok) { setCartStatus(body.error ?? 'Failed'); return }
      setCartStatus(body.status === 'ok'
        ? `All ${combo.items.length} items added to your cart ✓`
        : `Added ${body.addedItems?.length ?? 0} – ${body.skippedItems?.length ?? 0} out of stock`)
    } catch {
      setCartStatus('Cart service unavailable')
    } finally {
      setAddingToCart(false)
    }
  }

  if (loading) return <div className="spinner" style={{ marginTop: 80 }} />

  if (error) return (
    <div className="shared-page">
      <div className="error-banner">
        {error.includes('404') || error.includes('not found')
          ? '🔍 This combo link is no longer active or has been made private.'
          : `⚠️ ${error}`}
      </div>
    </div>
  )

  if (!combo) return null

  const shareUrl = window.location.href

  return (
    <div className="shared-page">
      <header className="shared-header">
        <p className="shared-branding">Combo Portfolio</p>
        <h1>{combo.name}</h1>
        <p>{combo.items.length} items · curated outfit</p>
      </header>

      {/* Item list */}
      <div className="item-list" style={{ padding: 0 }}>
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

      {/* Cart status banner */}
      {cartStatus && (
        <div style={{ margin: '16px 0', padding: '12px 16px', borderRadius: 8, background: '#e8f5e9', color: '#27ae60', fontWeight: 500, fontSize: 14 }}>
          {cartStatus}
        </div>
      )}

      {/* Footer actions */}
      <div className="detail-footer" style={{ position: 'static', marginTop: 24 }}>
        <button
          className="btn btn-primary btn-full btn-lg"
          onClick={handleAddAllToCart}
          disabled={addingToCart}
        >
          {addingToCart ? 'Adding to cart…' : `Add All ${combo.items.length} Items to Cart`}
        </button>

        <button
          className="btn btn-secondary btn-full"
          onClick={() => setShowShare(true)}
        >
          🔗 Share this combo
        </button>
      </div>

      <footer className="shared-footer" style={{ marginTop: 40 }}>
        <p className="shared-branding">Powered by Combo Portfolio</p>
      </footer>

      {showShare && (
        <ShareSheet
          shareUrl={shareUrl}
          comboName={combo.name}
          onClose={() => setShowShare(false)}
        />
      )}
    </div>
  )
}
