import { useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from './authStore'

export default function LoginPage() {
  const [username, setUsername] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const login = useAuthStore((s) => s.login)
  const navigate = useNavigate()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')

    if (!/^[a-zA-Z0-9]+$/.test(username) || username.length < 3 || username.length > 15) {
      setError('Username must be 3–15 letters or numbers, no spaces or symbols')
      return
    }

    setLoading(true)
    try {
      await login(username)
      navigate('/', { replace: true })
    } catch {
      setError('Could not sign in. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0f1a2e' }}>
      <div style={{ background: '#1a2a40', padding: '2rem', borderRadius: '12px', width: '100%', maxWidth: '360px', boxShadow: '0 8px 32px rgba(0,0,0,0.4)' }}>
        <h1 style={{ color: '#e2c97e', marginBottom: '0.25rem', textAlign: 'center', fontSize: '1.8rem' }}>♠ DeckForge</h1>
        <p style={{ color: '#7a9bb5', textAlign: 'center', marginBottom: '1.5rem', fontSize: '0.9rem' }}>Enter your username to play</p>

        <form onSubmit={handleSubmit}>
          <input
            type="text"
            placeholder="Username (e.g. alice42)"
            value={username}
            maxLength={15}
            onChange={(e) => setUsername(e.target.value)}
            style={{ width: '100%', padding: '0.75rem', borderRadius: '8px', border: '1px solid #2d4a6a', background: '#0f1a2e', color: '#e2e8f0', fontSize: '1rem', boxSizing: 'border-box' }}
            autoFocus
          />
          {error && <p style={{ color: '#f87171', fontSize: '0.85rem', marginTop: '0.5rem' }}>{error}</p>}
          <button
            type="submit"
            disabled={loading}
            style={{ width: '100%', marginTop: '1rem', padding: '0.75rem', borderRadius: '8px', background: '#e2c97e', color: '#0f1a2e', fontWeight: 700, fontSize: '1rem', border: 'none', cursor: loading ? 'not-allowed' : 'pointer', opacity: loading ? 0.7 : 1 }}
          >
            {loading ? 'Joining...' : 'Enter the Table Room'}
          </button>
        </form>

        <p style={{ color: '#4a6a8a', fontSize: '0.75rem', textAlign: 'center', marginTop: '1rem' }}>
          New username = new account. No password required.
        </p>
      </div>
    </div>
  )
}
