import { useState, useEffect, useCallback } from 'react';
import { fetchWishlist, removeFromWishlist } from '../api';
import { isLoggedIn } from '../auth';

const s = {
  page: { maxWidth: 800, margin: '0 auto', padding: '32px 24px' },
  title: { fontSize: 22, fontWeight: 700, marginBottom: 24 },
  unauthBox: {
    textAlign: 'center', padding: '60px 0', color: '#888',
  },
  unauthMsg: { fontSize: 16, marginBottom: 8 },
  unauthHint: { fontSize: 13 },
  empty: { textAlign: 'center', color: '#aaa', marginTop: 60, fontSize: 15 },
  list: { display: 'flex', flexDirection: 'column', gap: 16 },
  item: {
    background: '#fff',
    border: '1px solid #eee',
    borderRadius: 12,
    padding: '16px 20px',
    display: 'flex',
    alignItems: 'center',
    gap: 20,
  },
  itemImg: {
    width: 72, height: 96,
    background: '#f5f5f5',
    borderRadius: 8,
    display: 'flex', alignItems: 'center', justifyContent: 'center',
    fontSize: 36, flexShrink: 0,
  },
  itemInfo: { flex: 1 },
  itemBrand: { fontSize: 11, color: '#888', textTransform: 'uppercase', letterSpacing: 0.5, marginBottom: 3 },
  itemName: { fontSize: 15, fontWeight: 500, marginBottom: 4 },
  itemMeta: { fontSize: 13, color: '#555' },
  itemPrice: { fontSize: 15, fontWeight: 600, marginTop: 4 },
  outOfStock: { fontSize: 12, color: '#e44', marginLeft: 6 },
  removeBtn: {
    background: 'none',
    border: '1px solid #e44',
    color: '#e44',
    borderRadius: 6,
    padding: '6px 12px',
    fontSize: 13,
    flexShrink: 0,
  },
  total: { fontSize: 13, color: '#888', marginBottom: 16 },
};

const categoryEmoji = { tops: '👕', bottoms: '👖', dresses: '👗', outerwear: '🧥' };

function emojiForSku(configSku) {
  // Map by prefix category from known products
  const known = {
    'PD-001': '👕', 'PD-002': '👖', 'PD-003': '👗', 'PD-004': '🧥',
    'PD-005': '👕', 'PD-006': '👗', 'PD-007': '👕', 'PD-008': '👚',
  };
  return known[configSku] || '🛍️';
}

export default function WishlistPage({ navigate }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [unauthenticated, setUnauthenticated] = useState(false);
  const [removing, setRemoving] = useState(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchWishlist();
      if (data.unauthenticated) {
        setUnauthenticated(true);
        setItems([]);
      } else {
        setUnauthenticated(false);
        setItems(data.items ?? []);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  // Re-check when login state changes
  useEffect(() => {
    const id = setInterval(() => {
      if (isLoggedIn() && unauthenticated) load();
    }, 600);
    return () => clearInterval(id);
  }, [unauthenticated, load]);

  async function handleRemove(configSku) {
    setRemoving(configSku);
    try {
      await removeFromWishlist(configSku);
      setItems(prev => prev.filter(i => i.configSku !== configSku));
    } finally {
      setRemoving(null);
    }
  }

  if (loading) return <div style={{ ...s.page, color: '#aaa' }}>Loading wishlist…</div>;

  if (unauthenticated) {
    return (
      <div style={s.page}>
        <div style={s.title}>My Wishlist</div>
        <div style={s.unauthBox}>
          <div style={s.unauthMsg}>You are not logged in.</div>
          <div style={s.unauthHint}>Use the Login button in the nav to authenticate as the demo shopper.</div>
        </div>
      </div>
    );
  }

  return (
    <div style={s.page}>
      <div style={s.title}>My Wishlist</div>
      {items.length > 0 && (
        <div style={s.total}>{items.length} saved item{items.length !== 1 ? 's' : ''}</div>
      )}

      {items.length === 0 ? (
        <div style={s.empty}>
          Your wishlist is empty.{' '}
          <span
            style={{ color: '#1a1a1a', fontWeight: 500, cursor: 'pointer' }}
            onClick={() => navigate({ name: 'list' })}
          >
            Browse products →
          </span>
        </div>
      ) : (
        <div style={s.list}>
          {items.map(item => (
            <div key={item.itemId} style={s.item}>
              <div
                style={s.itemImg}
                onClick={() => navigate({ name: 'detail', configSku: item.configSku })}
              >
                {emojiForSku(item.configSku)}
              </div>
              <div style={s.itemInfo}>
                <div style={s.itemBrand}>{item.brand}</div>
                <div
                  style={{ ...s.itemName, cursor: 'pointer' }}
                  onClick={() => navigate({ name: 'detail', configSku: item.configSku })}
                >
                  {item.name}
                </div>
                <div style={s.itemMeta}>
                  Size: {item.size} &nbsp;·&nbsp; Colour: {item.color}
                </div>
                <div>
                  <span style={s.itemPrice}>${item.price?.amount?.toFixed(2)}</span>
                  {!item.inStock && <span style={s.outOfStock}>Out of stock</span>}
                </div>
              </div>
              <button
                style={s.removeBtn}
                onClick={() => handleRemove(item.configSku)}
                disabled={removing === item.configSku}
              >
                {removing === item.configSku ? '…' : 'Remove'}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
