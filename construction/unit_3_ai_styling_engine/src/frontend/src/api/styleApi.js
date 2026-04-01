// All requests use credentials: 'include' to forward the session cookie.
// The Vite dev server proxies /api/* → http://localhost:8080.

const BASE = '/api/v1/style'

async function handleResponse(res) {
  if (res.status === 403) throw new Error('UNAUTHENTICATED')
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

/**
 * GET /api/v1/style/preferences/options
 * @returns {{ occasions: string[], styles: string[], colors: string[] }}
 */
export async function getPreferenceOptions() {
  const res = await fetch(`${BASE}/preferences/options`, {
    credentials: 'include',
  })
  return handleResponse(res)
}

/**
 * POST /api/v1/style/preferences/confirm
 * @param {object} preferences
 * @returns {{ summary: string, preferences: object }}
 */
export async function confirmPreferences(preferences) {
  const res = await fetch(`${BASE}/preferences/confirm`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(preferences),
  })
  return handleResponse(res)
}

/**
 * POST /api/v1/style/combos/generate
 * @param {{ preferences?: object, excludeComboIds?: string[] }} body
 * @returns {object} success or fallback response
 */
export async function generateCombos(body = {}) {
  const res = await fetch(`${BASE}/combos/generate`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  return handleResponse(res)
}
