import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getPreferenceOptions, confirmPreferences, generateCombos } from '../api/styleApi'
import s from './screen.module.css'

export default function PreferenceInputScreen() {
  const navigate = useNavigate()
  const [options, setOptions] = useState({ occasions: [], styles: [], colors: [] })
  const [selectedOccasions, setSelectedOccasions] = useState([])
  const [selectedStyle, setSelectedStyle] = useState(null)
  const [budget, setBudget] = useState({ min: '', max: '' })
  const [preferredColors, setPreferredColors] = useState([])
  const [excludedColors, setExcludedColors] = useState([])
  const [freeText, setFreeText] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    getPreferenceOptions()
      .then(setOptions)
      .catch((e) => setError(e.message))
  }, [])

  function toggleChip(value, selected, setSelected) {
    setSelected((prev) =>
      prev.includes(value) ? prev.filter((v) => v !== value) : [...prev, value]
    )
  }

  async function handleSurpriseMe() {
    setLoading(true)
    setError(null)
    try {
      const result = await generateCombos({})
      navigate('/style/combos', { state: { result, excludeComboIds: [] } })
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  async function handleGenerate() {
    setLoading(true)
    setError(null)
    try {
      const prefs = {
        occasions: selectedOccasions,
        styles: selectedStyle ? [selectedStyle] : [],
        budget: budget.min !== '' && budget.max !== ''
          ? { min: Number(budget.min), max: Number(budget.max) }
          : undefined,
        colors: {
          preferred: preferredColors,
          excluded: excludedColors,
        },
        freeText,
      }
      const confirmation = await confirmPreferences(prefs)
      navigate('/style/confirm', { state: { confirmation, prefs } })
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={s.screen}>
      <header className={s.header}>
        <h1>Customise your combo</h1>
      </header>

      {error && <p className={s.error}>{error}</p>}

      <section>
        <h3>Occasion</h3>
        <div className={s.chips}>
          {options.occasions.map((o) => (
            <button
              key={o}
              className={`${s.chip} ${selectedOccasions.includes(o) ? s.chipActive : ''}`}
              onClick={() => toggleChip(o, selectedOccasions, setSelectedOccasions)}
            >
              {o}
            </button>
          ))}
        </div>
      </section>

      <section>
        <h3>Style direction</h3>
        <div className={s.chips}>
          {options.styles.map((style) => (
            <button
              key={style}
              className={`${s.chip} ${selectedStyle === style ? s.chipActive : ''}`}
              onClick={() => setSelectedStyle(style === selectedStyle ? null : style)}
            >
              {style}
            </button>
          ))}
        </div>
      </section>

      <section>
        <h3>Budget</h3>
        <div className={s.row}>
          <input
            type="number"
            placeholder="Min $"
            value={budget.min}
            onChange={(e) => setBudget((b) => ({ ...b, min: e.target.value }))}
            className={s.input}
          />
          <input
            type="number"
            placeholder="Max $"
            value={budget.max}
            onChange={(e) => setBudget((b) => ({ ...b, max: e.target.value }))}
            className={s.input}
          />
        </div>
      </section>

      <section>
        <h3>Colours I love</h3>
        <div className={s.chips}>
          {options.colors.map((c) => (
            <button
              key={c}
              className={`${s.chip} ${preferredColors.includes(c) ? s.chipActive : ''}`}
              onClick={() => toggleChip(c, preferredColors, setPreferredColors)}
            >
              {c}
            </button>
          ))}
        </div>
        <h3>Colours to avoid</h3>
        <div className={s.chips}>
          {options.colors.map((c) => (
            <button
              key={c}
              className={`${s.chip} ${excludedColors.includes(c) ? s.chipDanger : ''}`}
              onClick={() => toggleChip(c, excludedColors, setExcludedColors)}
            >
              {c}
            </button>
          ))}
        </div>
      </section>

      <section>
        <h3>Tell us more</h3>
        <textarea
          className={s.textarea}
          placeholder="e.g. something light for a summer trip"
          value={freeText}
          onChange={(e) => setFreeText(e.target.value)}
          rows={3}
        />
      </section>

      <footer className={s.footer}>
        <button className={s.btnSecondary} onClick={handleSurpriseMe} disabled={loading}>
          Surprise me
        </button>
        <button className={s.btnPrimary} onClick={handleGenerate} disabled={loading}>
          {loading ? 'Generating…' : 'Generate Combo →'}
        </button>
      </footer>
    </div>
  )
}
