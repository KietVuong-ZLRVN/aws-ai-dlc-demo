import { useState, useEffect, useCallback } from 'react';
import { fetchProducts, fetchWishlist } from '../api';
import ProductCard from '../components/ProductCard';

const s = {
  page: { display: 'flex', minHeight: 'calc(100vh - 56px)' },
  sidebar: {
    width: 220,
    flexShrink: 0,
    background: '#fff',
    borderRight: '1px solid #eee',
    padding: '24px 20px',
  },
  filterTitle: { fontSize: 11, fontWeight: 700, textTransform: 'uppercase', letterSpacing: 1, color: '#888', marginBottom: 10 },
  filterSection: { marginBottom: 24 },
  checkRow: { display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6, fontSize: 14, cursor: 'pointer' },
  main: { flex: 1, padding: '24px 28px' },
  topBar: { display: 'flex', gap: 12, marginBottom: 24, alignItems: 'center' },
  searchInput: {
    flex: 1,
    border: '1px solid #ddd',
    borderRadius: 8,
    padding: '9px 14px',
    fontSize: 14,
    outline: 'none',
  },
  priceInput: {
    border: '1px solid #ddd',
    borderRadius: 8,
    padding: '9px 12px',
    fontSize: 14,
    width: 130,
    outline: 'none',
  },
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
    gap: 20,
  },
  total: { fontSize: 13, color: '#888', marginBottom: 16 },
  empty: { textAlign: 'center', color: '#aaa', marginTop: 60, fontSize: 16 },
  error: { color: '#e44', padding: 20 },
};

export default function ProductListPage({ navigate, openLoginModal }) {
  const [products, setProducts] = useState([]);
  const [filters, setFilters] = useState(null);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState('');
  const [selectedColors, setSelectedColors] = useState([]);
  const [selectedCategory, setSelectedCategory] = useState('');
  const [price, setPrice] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [wishlistSkus, setWishlistSkus] = useState(new Set());

  const loadWishlist = useCallback(async () => {
    try {
      const data = await fetchWishlist();
      if (!data.unauthenticated) {
        setWishlistSkus(new Set(data.items?.map(i => i.configSku) ?? []));
      }
    } catch {
      // ignore wishlist errors on product list page
    }
  }, []);

  const loadProducts = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const params = { q: query, category: selectedCategory, colors: selectedColors };
      if (price) params.price = price;
      const data = await fetchProducts(params);
      setProducts(data.products ?? []);
      setTotal(data.total ?? 0);
      if (data.filters) setFilters(data.filters);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, [query, selectedCategory, selectedColors, price]);

  useEffect(() => { loadProducts(); }, [loadProducts]);
  useEffect(() => { loadWishlist(); }, [loadWishlist]);

  function toggleColor(c) {
    setSelectedColors(prev =>
      prev.includes(c) ? prev.filter(x => x !== c) : [...prev, c]
    );
  }

  return (
    <div style={s.page}>
      {/* Sidebar filters */}
      <aside style={s.sidebar}>
        {filters && (
          <>
            <div style={s.filterSection}>
              <div style={s.filterTitle}>Category</div>
              {['', ...filters.categories].map(cat => (
                <label key={cat || 'all'} style={s.checkRow}>
                  <input
                    type="radio"
                    name="category"
                    checked={selectedCategory === cat}
                    onChange={() => setSelectedCategory(cat)}
                  />
                  {cat || 'All'}
                </label>
              ))}
            </div>

            <div style={s.filterSection}>
              <div style={s.filterTitle}>Colour</div>
              {filters.colors.map(c => (
                <label key={c} style={s.checkRow}>
                  <input
                    type="checkbox"
                    checked={selectedColors.includes(c)}
                    onChange={() => toggleColor(c)}
                  />
                  {c.charAt(0).toUpperCase() + c.slice(1)}
                </label>
              ))}
            </div>
          </>
        )}
      </aside>

      {/* Main content */}
      <main style={s.main}>
        <div style={s.topBar}>
          <input
            style={s.searchInput}
            placeholder="Search products..."
            value={query}
            onChange={e => setQuery(e.target.value)}
          />
          <input
            style={s.priceInput}
            placeholder="Price: e.g. 0-60"
            value={price}
            onChange={e => setPrice(e.target.value)}
          />
        </div>

        {error && <div style={s.error}>{error}</div>}

        {!loading && (
          <div style={s.total}>{total} product{total !== 1 ? 's' : ''} found</div>
        )}

        {loading ? (
          <div style={s.empty}>Loading…</div>
        ) : products.length === 0 ? (
          <div style={s.empty}>No products found</div>
        ) : (
          <div style={s.grid}>
            {products.map(p => (
              <ProductCard
                key={p.configSku}
                product={p}
                onClick={() => navigate({ name: 'detail', configSku: p.configSku })}
                openLoginModal={openLoginModal}
                wishlisted={wishlistSkus.has(p.configSku)}
                onWishlistChange={loadWishlist}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
