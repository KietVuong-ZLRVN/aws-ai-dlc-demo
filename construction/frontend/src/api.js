// API clients for both backend services.
// Product Discovery runs on :8080, Wishlist on :8081.

const PD_BASE = 'http://localhost:8080/api/v1';
const WL_BASE = 'http://localhost:8081/api/v1';

// ─── Product Discovery ────────────────────────────────────────────────────────

export async function fetchProducts(params = {}) {
  const qs = new URLSearchParams();
  if (params.q) qs.set('q', params.q);
  if (params.category) qs.set('category', params.category);
  if (params.colors?.length) params.colors.forEach(c => qs.append('colors', c));
  if (params.price) qs.set('price', params.price);
  if (params.offset != null) qs.set('offset', params.offset);
  if (params.limit != null) qs.set('limit', params.limit);
  const res = await fetch(`${PD_BASE}/products?${qs}`);
  if (!res.ok) throw new Error(`Product list failed: ${res.status}`);
  return res.json();
}

export async function fetchProductDetail(configSku) {
  const res = await fetch(`${PD_BASE}/products/${configSku}`);
  if (res.status === 404) return null;
  if (!res.ok) throw new Error(`Product detail failed: ${res.status}`);
  return res.json();
}

// ─── Wishlist ─────────────────────────────────────────────────────────────────

export async function fetchWishlist() {
  const res = await fetch(`${WL_BASE}/wishlist`, { credentials: 'include' });
  if (res.status === 403) return { unauthenticated: true };
  if (!res.ok) throw new Error(`Wishlist fetch failed: ${res.status}`);
  return res.json();
}

export async function addToWishlist(simpleSku) {
  const res = await fetch(`${WL_BASE}/wishlist/items`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ simpleSku }),
  });
  const data = await res.json().catch(() => ({}));
  if (res.status === 403) return { requiresAuth: true, ...data };
  if (res.status === 409) return { alreadyPresent: true };
  if (!res.ok) throw new Error(`Add to wishlist failed: ${res.status}`);
  return data;
}

export async function removeFromWishlist(configSku) {
  const res = await fetch(`${WL_BASE}/wishlist/items/${configSku}`, {
    method: 'DELETE',
    credentials: 'include',
  });
  if (res.status === 403) return { requiresAuth: true };
  if (!res.ok) throw new Error(`Remove from wishlist failed: ${res.status}`);
  return {};
}
