import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { MyCombosPage } from './pages/MyCombosPage'
import { ComboDetailPage } from './pages/ComboDetailPage'
import { SharedComboPage } from './pages/SharedComboPage'
import { ToastProvider } from './components/Toast'

export default function App() {
  return (
    <BrowserRouter>
      <ToastProvider>
        <Routes>
          {/* My Combos list */}
          <Route path="/" element={<MyCombosPage />} />

          {/* Combo detail (authenticated) */}
          <Route path="/combos/:id" element={<ComboDetailPage />} />

          {/* Public shared combo — no login required */}
          <Route path="/shared/:token" element={<SharedComboPage />} />

          {/* Catch-all */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </ToastProvider>
    </BrowserRouter>
  )
}
