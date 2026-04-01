import { useNavigate, useLocation } from 'react-router-dom'
import { useState } from 'react'
import { generateCombos } from '../api/styleApi'
import s from './screen.module.css'

export default function ConfirmationScreen() {
  const navigate = useNavigate()
  const { state } = useLocation()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  if (!state?.confirmation) {
    navigate('/style/preferences', { replace: true })
    return null
  }

  const { confirmation, prefs } = state

  async function handleConfirm() {
    setLoading(true)
    setError(null)
    try {
      const result = await generateCombos({ preferences: prefs, excludeComboIds: [] })
      navigate('/style/combos', { state: { result, prefs, excludeComboIds: [] } })
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={s.screen}>
      <header className={s.header}>
        <button className={s.backBtn} onClick={() => navigate(-1)}>← Back</button>
        <h1>Does this sound right?</h1>
      </header>

      <div className={s.summaryCard}>
        <p>{confirmation.summary}</p>
      </div>

      {error && <p className={s.error}>{error}</p>}

      <footer className={s.footer}>
        <button className={s.btnSecondary} onClick={() => navigate('/style/preferences')}>
          Edit preferences
        </button>
        <button className={s.btnSecondary} onClick={() => navigate('/')}>
          Cancel
        </button>
        <button className={s.btnPrimary} onClick={handleConfirm} disabled={loading}>
          {loading ? 'Generating…' : 'Looks good, generate!'}
        </button>
      </footer>
    </div>
  )
}
