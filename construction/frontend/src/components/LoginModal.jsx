import { login } from '../auth';

const s = {
  overlay: {
    position: 'fixed', inset: 0,
    background: 'rgba(0,0,0,0.45)',
    display: 'flex', alignItems: 'center', justifyContent: 'center',
    zIndex: 200,
  },
  modal: {
    background: '#fff',
    borderRadius: 12,
    padding: 32,
    width: 360,
    maxWidth: '90vw',
    boxShadow: '0 20px 60px rgba(0,0,0,0.2)',
  },
  title: { fontSize: 18, fontWeight: 700, marginBottom: 8 },
  body: { fontSize: 14, color: '#555', marginBottom: 24, lineHeight: 1.5 },
  loginBtn: {
    width: '100%',
    background: '#1a1a1a',
    color: '#fff',
    border: 'none',
    borderRadius: 8,
    padding: '12px 0',
    fontSize: 15,
    fontWeight: 600,
    marginBottom: 10,
  },
  cancelBtn: {
    width: '100%',
    background: 'none',
    border: '1px solid #ddd',
    borderRadius: 8,
    padding: '11px 0',
    fontSize: 14,
    color: '#555',
  },
  skuHint: {
    fontSize: 12, color: '#888', marginBottom: 16,
    background: '#f5f5f5', borderRadius: 6, padding: '6px 10px',
  },
};

export default function LoginModal({ pendingSku, onSuccess, onClose }) {
  function handleLogin() {
    login();
    onSuccess();
  }

  return (
    <div style={s.overlay} onClick={onClose}>
      <div style={s.modal} onClick={e => e.stopPropagation()}>
        <p style={s.title}>Login required</p>
        <p style={s.body}>
          You need to be logged in to add items to your wishlist.
        </p>
        {pendingSku && (
          <p style={s.skuHint}>
            After login, <strong>{pendingSku}</strong> will be added automatically.
          </p>
        )}
        <button style={s.loginBtn} onClick={handleLogin}>
          Login as shopper-123 (demo)
        </button>
        <button style={s.cancelBtn} onClick={onClose}>Cancel</button>
      </div>
    </div>
  );
}
