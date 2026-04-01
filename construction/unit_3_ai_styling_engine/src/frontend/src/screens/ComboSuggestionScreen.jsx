import { useNavigate, useLocation } from 'react-router-dom'
import { useState } from 'react'
import { generateCombos } from '../api/styleApi'
import s from './screen.module.css'

export default function ComboSuggestionScreen() {
  const navigate = useNavigate()
  const { state } = useLocation()
  const [result, setResult] = useState(state?.result)
  const [excludeComboIds, setExcludeComboIds] = useState(state?.excludeComboIds ?? [])
  const [prefs] = useState(state?.prefs)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  if (!result) {
    navigate('/style/preferences', { replace: true })
    return null
  }

  // ── Fallback screen ────────────────────────────────────────────────────────
  if (result.status === 'fallback') {
    return (
      <div className={s.screen}>
        <header className={s.header}>
          <h1>Let's adjust your wishlist</h1>
        </header>
        <p className={s.message}>{result.message}</p>

        <div className={s.cardList}>
          {result.alternatives.map((alt) => (
            <div key={alt.configSku} className={s.card}>
              <img src={alt.imageUrl} alt={alt.name} className={s.productImage} onError={(e) => { e.target.style.display = 'none' }} />
              <div className={s.cardBody}>
                <p className={s.productName}>{alt.name}</p>
                <p className={s.brand}>{alt.brand}</p>
                <p className={s.price}>${alt.price.toFixed(2)}</p>
                <p className={s.reason}>{alt.reason}</p>
              </div>
            </div>
          ))}
        </div>

        <footer className={s.footer}>
          <button className={s.btnSecondary} onClick={() => navigate('/style/preferences')}>
            Edit preferences
          </button>
        </footer>
      </div>
    )
  }

  // ── Success screen ─────────────────────────────────────────────────────────
  function rejectCombo(comboId) {
    setExcludeComboIds((prev) => [...prev, comboId])
  }

  async function handleGenerateNew() {
    setLoading(true)
    setError(null)
    try {
      const newResult = await generateCombos({ preferences: prefs, excludeComboIds })
      setResult(newResult)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  const visibleCombos = result.combos?.filter((c) => !excludeComboIds.includes(c.id)) ?? []

  return (
    <div className={s.screen}>
      <header className={s.header}>
        <h1>Your combos</h1>
      </header>

      {error && <p className={s.error}>{error}</p>}

      {loading && <p className={s.loading}>Generating new combos…</p>}

      <div className={s.cardList}>
        {visibleCombos.map((combo) => (
          <div key={combo.id} className={s.comboCard}>
            <div className={s.productRow}>
              {combo.items.map((item) => (
                <div key={item.simpleSku} className={s.productTile}>
                  <img src={item.imageUrl} alt={item.name} className={s.productImage} onError={(e) => { e.target.style.display = 'none' }} />
                  {item.source === 'catalog' && (
                    <span className={s.catalogBadge}>Suggested for you</span>
                  )}
                  <p className={s.productName}>{item.name}</p>
                  <p className={s.brand}>{item.brand}</p>
                  <p className={s.price}>${item.price.toFixed(2)}</p>
                </div>
              ))}
            </div>

            <p className={s.reasoning}>{combo.reasoning}</p>

            <div className={s.comboActions}>
              <button
                className={s.btnDanger}
                onClick={() => rejectCombo(combo.id)}
              >
                Not for me
              </button>
            </div>
          </div>
        ))}
      </div>

      {result.exhausted ? (
        <p className={s.exhaustedMsg}>
          We've shown you all available combos. Try adjusting your preferences.
        </p>
      ) : (
        <footer className={s.footer}>
          <button className={s.btnSecondary} onClick={() => navigate('/style/preferences')}>
            Edit preferences
          </button>
          <button className={s.btnPrimary} onClick={handleGenerateNew} disabled={loading}>
            Generate new combos
          </button>
        </footer>
      )}
    </div>
  )
}
