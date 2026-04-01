import { useState } from 'react'

interface ShareSheetProps {
  shareUrl: string
  comboName: string
  onClose: () => void
}

export function ShareSheet({ shareUrl, comboName, onClose }: ShareSheetProps) {
  const [copied, setCopied] = useState(false)

  function handleCopy() {
    navigator.clipboard.writeText(shareUrl).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  const encodedUrl = encodeURIComponent(shareUrl)
  const encodedText = encodeURIComponent(`Check out my outfit combo: ${comboName}`)

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2 className="modal-title">Share "{comboName}"</h2>

        {/* Share URL display */}
        <div className="share-url-box">
          <span className="share-url-text">{shareUrl}</span>
          <button className="btn btn-secondary" style={{ padding: '6px 12px', fontSize: 12 }} onClick={handleCopy}>
            {copied ? '✓ Copied' : 'Copy'}
          </button>
        </div>

        {/* Social actions */}
        <div className="share-actions">
          <button className="share-action-btn" onClick={handleCopy}>
            <span className="share-action-icon">🔗</span>
            {copied ? 'Link copied!' : 'Copy link'}
          </button>

          <a
            className="share-action-btn"
            href={`https://wa.me/?text=${encodedText}%20${encodedUrl}`}
            target="_blank"
            rel="noopener noreferrer"
          >
            <span className="share-action-icon">💬</span>
            Share to WhatsApp
          </a>

          <a
            className="share-action-btn"
            href={`https://www.instagram.com/`}
            target="_blank"
            rel="noopener noreferrer"
          >
            <span className="share-action-icon">📸</span>
            Share to Instagram
          </a>
        </div>

        <p className="share-note">Anyone with this link can view your combo.</p>

        <div className="modal-actions" style={{ marginTop: 20 }}>
          <button className="btn btn-secondary btn-full" onClick={onClose}>Close</button>
        </div>
      </div>
    </div>
  )
}
