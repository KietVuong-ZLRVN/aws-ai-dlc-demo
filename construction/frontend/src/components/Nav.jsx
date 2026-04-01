import { useState, useEffect } from 'react';
import { isLoggedIn, login, logout } from '../auth';

const s = {
  nav: {
    background: '#fff',
    borderBottom: '1px solid #e5e5e5',
    padding: '0 24px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    height: 56,
    position: 'sticky',
    top: 0,
    zIndex: 100,
  },
  logo: { fontWeight: 700, fontSize: 18, letterSpacing: '-0.5px', cursor: 'pointer' },
  links: { display: 'flex', gap: 24, alignItems: 'center' },
  link: (active) => ({
    padding: '4px 0',
    borderBottom: active ? '2px solid #1a1a1a' : '2px solid transparent',
    fontWeight: active ? 600 : 400,
    cursor: 'pointer',
    fontSize: 14,
  }),
  loginBtn: {
    background: '#1a1a1a',
    color: '#fff',
    border: 'none',
    borderRadius: 6,
    padding: '6px 14px',
    fontSize: 13,
    fontWeight: 500,
  },
  logoutBtn: {
    background: 'none',
    border: '1px solid #ccc',
    borderRadius: 6,
    padding: '5px 13px',
    fontSize: 13,
  },
  badge: {
    background: '#f0f0f0',
    borderRadius: 10,
    padding: '1px 7px',
    fontSize: 12,
    marginLeft: 4,
  },
};

export default function Nav({ page, navigate }) {
  const [loggedIn, setLoggedIn] = useState(isLoggedIn());

  // Re-check cookie state on every render (crude but sufficient for demo).
  useEffect(() => {
    const id = setInterval(() => setLoggedIn(isLoggedIn()), 500);
    return () => clearInterval(id);
  }, []);

  function handleLogin() {
    login();
    setLoggedIn(true);
  }

  function handleLogout() {
    logout();
    setLoggedIn(false);
  }

  return (
    <nav style={s.nav}>
      <span style={s.logo} onClick={() => navigate({ name: 'list' })}>ZALORA Demo</span>
      <div style={s.links}>
        <span
          style={s.link(page.name === 'list')}
          onClick={() => navigate({ name: 'list' })}
        >
          Products
        </span>
        <span
          style={s.link(page.name === 'wishlist')}
          onClick={() => navigate({ name: 'wishlist' })}
        >
          Wishlist
        </span>
        {loggedIn ? (
          <button style={s.logoutBtn} onClick={handleLogout}>
            Logout <span style={s.badge}>shopper-123</span>
          </button>
        ) : (
          <button style={s.loginBtn} onClick={handleLogin}>Login</button>
        )}
      </div>
    </nav>
  );
}
