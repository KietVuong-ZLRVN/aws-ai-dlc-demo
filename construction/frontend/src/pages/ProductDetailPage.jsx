import { useState, useEffect } from 'react';
import { fetchProductDetail, addToWishlist } from '../api';
import { isLoggedIn } from '../auth';

const s = {
  page: { maxWidth: 900, margin: '0 auto', padding: '32px 24px' },
  back: { fontSize: 13, color: '#888', cursor: 'pointer', marginBottom: 24, display: 'inline-flex', alignItems: 'center', gap: 4 },
  layout: { display: 'flex', gap: 48, flexWrap: 'wrap' },
  imgBox: {
    width: 340, flexShrink: 0,
    background: '#f5f5f5',
    borderRadius: 12,
    aspectRatio: '3/4',
    display: 'flex', alignItems: 'center', justifyContent: 'center',
    fontSize: 80,
  },
  info: { flex: 1, minWidth: 260 },
  brand: { fontSize: 12, color: '#888', textTransform: 'uppercase', letterSpacing: 1, marginBottom: 6 },
  name: { fontSize: 26, fontWeight: 600, marginBottom: 12, lineHeight: 1.2 },
  price: { fontSize: 22, fontWeight: 700, marginBottom: 20 },
  section: { marginBottom: 20 },
  label: { fontSize: 12, fontWeight: 700, textTransform: 'uppercase', letterSpacing: 0.5, color: '#888', marginBottom: 8 },
  variantGrid: { display: 'flex', flexWrap: 'wrap', gap: 8 },
  variantBtn: (selected, inStock) => ({
    padding: '7px 14px',
    borderRadius: 6,
    border: selected ? '2px solid #1a1a1a' : '1px solid #ddd',
    background: selected ? '#1a1a1a' : '#fff',
    color: selected ? '#fff' : inStock ? '#1a1a1a' : '#ccc',
    fontSize: 13,
    fontWeight: selected ? 600 : 400,
    cursor: inStock ? 'pointer' : 'default',
    textDecoration: inStock ? 'none' : 'line-through',
  }),
  addBtn: (disabled) => ({
    width: '100%',
    background: disabled ? '#e0e0e0' : '#1a1a1a',
    color: disabled ? '#aaa' : '#fff',
    border: 'none',
    borderRadius: 10,
    padding: '14px 0',
    fontSize: 15,
    fontWeight: 600,
    marginTop: 8,
    cursor: disabled ? 'default' : 'pointer',
  }),
  toast: (type) => ({
    marginTop: 10,
    padding: '10px 14px',
    borderRadius: 8,
    fontSize: 13,
    background: type === 'success' ? '#ecfdf5' : type === 'error' ? '#fef2f2' : '#fff7ed',
    color: type === 'success' ? '#166534' : type === 'error' ? '#991b1b' : '#92400e',
    border: `1px solid ${type === 'success' ? '#bbf7d0' : type === 'error' ? '#fecaca' : '#fed7aa'}`,
  }),
  occasions: { display: 'flex', gap: 6, flexWrap: 'wrap' },
  tag: {
    background: '#f5f5f5', borderRadius: 20,
    padding: '3px 10px', fontSize: 12, color: '#555',
  },
  empty: { textAlign: 'center', color: '#aaa', marginTop: 80, fontSize: 16 },
};

const categoryEmoji = { tops: '👕', bottoms: '👖', dresses: '👗', outerwear: '🧥' };

export default function ProductDetailPage({ configSku, navigate, openLoginModal }) {
  const [product, setProduct] = useState(null);
  const [selected, setSelected] = useState(null); // selected variant SimpleSku
  const [loading, setLoading] = useState(true);
  const [adding, setAdding] = useState(false);
  const [toast, setToast] = useState(null);

  useEffect(() => {
    fetchProductDetail(configSku).then(p => {
      setProduct(p);
      if (p?.variants?.length) {
        const first = p.variants.find(v => v.inStock) ?? p.variants[0];
        setSelected(first.simpleSku);
      }
    }).finally(() => setLoading(false));
  }, [configSku]);

  async function handleAdd() {
    if (!selected) return;
    if (!isLoggedIn()) {
      openLoginModal(selected, () => doAdd(selected));
      return;
    }
    await doAdd(selected);
  }

  async function doAdd(simpleSku) {
    setAdding(true);
    setToast(null);
    try {
      const res = await addToWishlist(simpleSku);
      if (res.requiresAuth) {
        openLoginModal(simpleSku, () => doAdd(simpleSku));
        return;
      }
      if (res.alreadyPresent) {
        setToast({ type: 'warn', msg: 'This product is already in your wishlist.' });
      } else {
        setToast({ type: 'success', msg: 'Added to wishlist!' });
      }
    } catch {
      setToast({ type: 'error', msg: 'Could not add to wishlist. Try again.' });
    } finally {
      setAdding(false);
    }
  }

  if (loading) return <div style={s.empty}>Loading…</div>;
  if (!product) return <div style={s.empty}>Product not found.</div>;

  const emoji = categoryEmoji[product.configSku?.split('-')[0]?.toLowerCase()] || '🛍️';

  return (
    <div style={s.page}>
      <span style={s.back} onClick={() => navigate({ name: 'list' })}>
        ← Back to products
      </span>

      <div style={s.layout}>
        <div style={s.imgBox}>{emoji}</div>

        <div style={s.info}>
          <div style={s.brand}>{product.brand}</div>
          <div style={s.name}>{product.name}</div>
          <div style={s.price}>${product.price?.amount?.toFixed(2)}</div>

          <div style={s.section}>
            <div style={s.label}>
              Select size
              {selected && ` — ${product.variants.find(v => v.simpleSku === selected)?.size}`}
            </div>
            <div style={s.variantGrid}>
              {product.variants?.map(v => (
                <button
                  key={v.simpleSku}
                  style={s.variantBtn(selected === v.simpleSku, v.inStock)}
                  onClick={() => v.inStock && setSelected(v.simpleSku)}
                >
                  {v.size}
                </button>
              ))}
            </div>
          </div>

          {product.occasions?.length > 0 && (
            <div style={s.section}>
              <div style={s.label}>Occasions</div>
              <div style={s.occasions}>
                {product.occasions.map(o => <span key={o} style={s.tag}>{o}</span>)}
              </div>
            </div>
          )}

          <button
            style={s.addBtn(!selected || adding)}
            onClick={handleAdd}
            disabled={!selected || adding}
          >
            {adding ? 'Adding…' : '♡  Add to Wishlist'}
          </button>

          {toast && (
            <div style={s.toast(toast.type)}>{toast.msg}</div>
          )}
        </div>
      </div>
    </div>
  );
}
