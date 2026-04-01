import { useState } from 'react';
import Nav from './components/Nav';
import ProductListPage from './pages/ProductListPage';
import ProductDetailPage from './pages/ProductDetailPage';
import WishlistPage from './pages/WishlistPage';
import LoginModal from './components/LoginModal';

// Simple client-side router via a state machine (no router library needed).
export default function App() {
  const [page, setPage] = useState({ name: 'list' });
  const [loginModal, setLoginModal] = useState(null); // { pendingSku, onSuccess }

  function navigate(p) { setPage(p); }

  function openLoginModal(pendingSku, onSuccess) {
    setLoginModal({ pendingSku, onSuccess });
  }

  function closeLoginModal() { setLoginModal(null); }

  return (
    <>
      <Nav page={page} navigate={navigate} />

      {page.name === 'list' && (
        <ProductListPage navigate={navigate} openLoginModal={openLoginModal} />
      )}
      {page.name === 'detail' && (
        <ProductDetailPage
          configSku={page.configSku}
          navigate={navigate}
          openLoginModal={openLoginModal}
        />
      )}
      {page.name === 'wishlist' && (
        <WishlistPage navigate={navigate} />
      )}

      {loginModal && (
        <LoginModal
          pendingSku={loginModal.pendingSku}
          onSuccess={() => { closeLoginModal(); loginModal.onSuccess?.(); }}
          onClose={closeLoginModal}
        />
      )}
    </>
  );
}
