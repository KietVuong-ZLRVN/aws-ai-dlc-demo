import type { Combo, SaveComboPayload } from '../types/combo'

const BASE = '/api/v1'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    credentials: 'include', // session cookie
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error ?? res.statusText)
  }
  // 204 No Content
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export const comboApi = {
  list: (): Promise<Combo[]> =>
    request('/combos'),

  get: (id: string): Promise<Combo> =>
    request(`/combos/${id}`),

  save: (payload: SaveComboPayload): Promise<{ id: string }> =>
    request('/combos', { method: 'POST', body: JSON.stringify(payload) }),

  rename: (id: string, name: string): Promise<void> =>
    request(`/combos/${id}`, { method: 'PUT', body: JSON.stringify({ name }) }),

  setVisibility: (id: string, visibility: 'public' | 'private'): Promise<void> =>
    request(`/combos/${id}`, { method: 'PUT', body: JSON.stringify({ visibility }) }),

  remove: (id: string): Promise<void> =>
    request(`/combos/${id}`, { method: 'DELETE' }),

  share: (id: string): Promise<{ shareToken: string; shareUrl: string }> =>
    request(`/combos/${id}/share`, { method: 'POST' }),

  getShared: (token: string): Promise<Combo> =>
    request(`/combos/shared/${token}`),
}
