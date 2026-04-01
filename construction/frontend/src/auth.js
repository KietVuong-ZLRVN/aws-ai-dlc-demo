// Auth helpers for the demo mock session.
// The Wishlist backend recognises cookie  session=demo-session-token  as authenticated.

const SESSION_COOKIE = 'session';
const DEMO_TOKEN = 'demo-session-token';

export function isLoggedIn() {
  return document.cookie.split(';').some(c => c.trim().startsWith(`${SESSION_COOKIE}=`));
}

export function login() {
  // Set a cookie that the in-memory AuthSessionService will accept.
  // SameSite=Lax works for localhost cross-port requests when credentials:include is used.
  document.cookie = `${SESSION_COOKIE}=${DEMO_TOKEN}; path=/; SameSite=None; Secure`;
  // Fallback for http localhost (Secure is ignored on http in most browsers but SameSite=None requires it)
  document.cookie = `${SESSION_COOKIE}=${DEMO_TOKEN}; path=/`;
}

export function logout() {
  document.cookie = `${SESSION_COOKIE}=; path=/; max-age=0`;
}
