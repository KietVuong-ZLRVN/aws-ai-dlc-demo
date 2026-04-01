import { useState } from 'react';
import { addToWishlist } from '../api';
import { isLoggedIn } from '../auth';

const s = {
  card: {
    background: '#fff',
    border: '1px solid #eee',
    borderRadius: 10,
    overflow: 'hidden',
    cursor: 'pointer',
    transition: 'box-shadow 0.15s',
    position: 'relative',
  },
  img: {
    width: '100%',
    aspectRatio: '3/4',
    background: '#f0f0f0',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: 48,
    userSelect: 'none',
  },
  info: { padding: '12px 14px 14px' },
  brand: { fontSize: 11, color: '#888', textTransform: 'uppercase', letterSpacing: 0.5 },
  name: { fontSize: 14, fontWeight: 500, margin: '2px 0 6px', lineHeight: 1.3 },
  price: { fontSize: 14, fontWeight: 600 },
  outOfStock: { fontSize: 12, color: '#e44', marginLeft: 6 },
  heartBtn: (wishlisted) => ({
    position: 'absolute',
    top: 10, right: 10,
    background: wishlisted ? '#ff4d6d' : 'rgba(255,255,255,0.85)',
    border: wishlisted ? 'none' : '1px solid #ddd',
    borderRadius: '50%',
    width: 34, height: 34,
    display: 'flex', alignItems: 'center', justifyContent: 'center',
    fontSize: 16,
    boxShadow: '0 1px 4px rgba(0,0,0,0.1)',
    color: wishlisted ? '#fff' : '#999',
    transition: 'all 0.15s',
  }),
  colors: {
    display: 'flex', gap: 4, marginTop: 6, flexWrap: 'wrap',
  },
  colorDot: (color) => ({
    width: 12, height: 12, borderRadius: '50%',
    background: colorMap[color] || color,
    border: '1px solid #ddd',
  }),
};

const colorMap = {
  white: '#f5f5f5',
  black: '#1a1a1a',
  blue: '#4a7fc1',
  navy: '#1f305e',
  pink: '#f4a7b9',
  red: '#e44',
  beige: '#d2b48c',
};

// Use a simple emoji as placeholder image based on category
const categoryEmoji = {
  tops: '👕', bottoms: '👖', dresses: '👗', outerwear: '🧥',
};

export default function ProductCard({ product, onClick, openLoginModal, wishlisted, onWishlistChange }) {
  const [loading, setLoading] = useState(false);

  async function handleHeart(e) {
    e.stopPropagation();
    if (!isLoggedIn()) {
      // Find a simpleSku to add — use first variant
      const sku = product.configSku + '-S'; // best-guess; detail page handles proper sku selection
      openLoginModal(sku, async () => {
        // After login, retry
        await doAdd(sku);
      });
      return;
    }
    const sku = product.configSku + '-S';
    await doAdd(sku);
  }

  async function doAdd(simpleSku) {
    setLoading(true);
    try {
      const res = await addToWishlist(simpleSku);
      if (res.requiresAuth) {
        openLoginModal(simpleSku, () => doAdd(simpleSku));
        return;
      }
      onWishlistChange?.();
    } finally {
      setLoading(false);
    }
  }

  const emoji = categoryEmoji[product.category] || '🛍️';

  return (
    <div style={s.card} onClick={onClick}>
      <div style={s.img}>{emoji}</div>
      <button
        style={s.heartBtn(wishlisted)}
        onClick={handleHeart}
        title={wishlisted ? 'In wishlist' : 'Add to wishlist'}
        disabled={loading}
      >
        {wishlisted ? '♥' : '♡'}
      </button>
      <div style={s.info}>
        <div style={s.brand}>{product.brand}</div>
        <div style={s.name}>{product.name}</div>
        <div>
          <span style={s.price}>${product.price?.amount?.toFixed(2)}</span>
          {!product.inStock && <span style={s.outOfStock}>Out of stock</span>}
        </div>
        <div style={s.colors}>
          {product.colors?.map(c => <span key={c} style={s.colorDot(c)} title={c} />)}
        </div>
      </div>
    </div>
  );
}
