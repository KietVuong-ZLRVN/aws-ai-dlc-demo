import { useState, useCallback, createContext, useContext } from 'react'

// ── Toast context ──────────────────────────────────────────────────────────

interface ToastCtx { showToast: (msg: string) => void }

const ToastContext = createContext<ToastCtx>({ showToast: () => {} })

export function useToast() {
  return useContext(ToastContext)
}

// ── Toast provider + renderer ──────────────────────────────────────────────

interface ToastMessage { id: number; text: string }

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<ToastMessage[]>([])
  let seq = 0

  const showToast = useCallback((text: string) => {
    const id = ++seq
    setToasts((prev) => [...prev, { id, text }])
    setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== id)), 2400)
  }, [])

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      <div className="toast-container">
        {toasts.map((t) => (
          <div key={t.id} className="toast">{t.text}</div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}
