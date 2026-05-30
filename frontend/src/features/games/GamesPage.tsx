import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useRequireAuth } from '../../shared/hooks/useRequireAuth'
import { useAuthStore } from '../auth/authStore'
import { useGames, useCreateGame, useJoinGame } from './useGames'

const s = {
  page: { minHeight: '100vh', background: '#0f1a2e', color: '#e2e8f0', padding: '2rem' },
  header: { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' },
  title: { color: '#e2c97e', fontSize: '1.6rem', fontWeight: 700 },
  btn: { padding: '0.6rem 1.2rem', borderRadius: '8px', border: 'none', cursor: 'pointer', fontWeight: 600 },
  btnGold: { background: '#e2c97e', color: '#0f1a2e' },
  btnOutline: { background: 'transparent', color: '#e2c97e', border: '1px solid #e2c97e' },
  grid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))', gap: '1rem' },
  card: { background: '#1a2a40', borderRadius: '12px', padding: '1.25rem', border: '1px solid #2d4a6a' },
  badge: { display: 'inline-block', padding: '0.2rem 0.6rem', borderRadius: '4px', fontSize: '0.75rem', fontWeight: 600 },
  modal: { position: 'fixed' as const, inset: 0, background: 'rgba(0,0,0,0.7)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 },
  modalBox: { background: '#1a2a40', borderRadius: '12px', padding: '2rem', width: '100%', maxWidth: '380px' },
  input: { width: '100%', padding: '0.6rem', borderRadius: '6px', border: '1px solid #2d4a6a', background: '#0f1a2e', color: '#e2e8f0', boxSizing: 'border-box' as const },
  label: { color: '#7a9bb5', fontSize: '0.85rem', display: 'block', marginBottom: '0.25rem', marginTop: '0.75rem' },
}

function statusColor(status: string) {
  if (status === 'WAITING') return { background: '#1a4a2a', color: '#4ade80' }
  if (status === 'IN_PROGRESS') return { background: '#1a2a4a', color: '#60a5fa' }
  return { background: '#2a1a1a', color: '#f87171' }
}

export default function GamesPage() {
  const authed = useRequireAuth()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const navigate = useNavigate()
  const { data: games, isLoading } = useGames()
  const createGame = useCreateGame()
  const joinGame = useJoinGame()
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState({ deckCount: 2, minPlayers: 2, maxPlayers: 8 })

  if (!authed) return null

  return (
    <div style={s.page}>
      <div style={s.header}>
        <span style={s.title}>♠ DeckForge</span>
        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          <span style={{ color: '#7a9bb5' }}>👤 {user?.username}</span>
          <button style={{ ...s.btn, ...s.btnGold }} onClick={() => setShowModal(true)}>+ New Table</button>
          <button style={{ ...s.btn, ...s.btnOutline }} onClick={() => { logout(); navigate('/login') }}>Logout</button>
        </div>
      </div>

      <h2 style={{ color: '#7a9bb5', marginBottom: '1rem', fontWeight: 400 }}>Open Tables</h2>

      {isLoading && <p style={{ color: '#4a6a8a' }}>Loading tables...</p>}
      {!isLoading && (!games || games.length === 0) && (
        <p style={{ color: '#4a6a8a' }}>No open tables yet. Create one!</p>
      )}

      <div style={s.grid}>
        {games?.filter(g => g.status !== 'FINISHED').map(game => (
          <div key={game.id} style={s.card}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
              <span style={{ color: '#e2c97e', fontWeight: 600 }}>🎩 {game.dealerUsername}</span>
              <span style={{ ...s.badge, ...statusColor(game.status) }}>{game.status}</span>
            </div>
            <div style={{ color: '#4a6a8a', fontSize: '0.75rem', marginBottom: '0.75rem' }}>
              {new Date(game.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </div>
            <div style={{ color: '#7a9bb5', fontSize: '0.85rem', lineHeight: 1.8 }}>
              <div>🃏 {game.deckCount} deck{game.deckCount !== 1 ? 's' : ''} · {game.deckCount * 52} cards</div>
              <div>👥 {game.playerCount} / {game.maxPlayers} players</div>
            </div>
            <button
              style={{ ...s.btn, ...s.btnGold, width: '100%', marginTop: '1rem' }}
              onClick={() => joinGame.mutate(game.id)}
            >
              Join Table
            </button>
          </div>
        ))}
      </div>

      {showModal && (
        <div style={s.modal} onClick={() => setShowModal(false)}>
          <div style={s.modalBox} onClick={e => e.stopPropagation()}>
            <h2 style={{ color: '#e2c97e', marginBottom: '1rem' }}>Create New Table</h2>
            <label style={s.label}>Number of Decks (1–8)</label>
            <input type="number" min={1} max={8} value={form.deckCount} style={s.input}
              onChange={e => setForm({ ...form, deckCount: +e.target.value })} />
            <label style={s.label}>Min Players</label>
            <input type="number" min={2} max={form.maxPlayers} value={form.minPlayers} style={s.input}
              onChange={e => setForm({ ...form, minPlayers: +e.target.value })} />
            <label style={s.label}>Max Players</label>
            <input type="number" min={form.minPlayers} max={8} value={form.maxPlayers} style={s.input}
              onChange={e => setForm({ ...form, maxPlayers: +e.target.value })} />
            <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.5rem' }}>
              <button style={{ ...s.btn, ...s.btnOutline, flex: 1 }} onClick={() => setShowModal(false)}>Cancel</button>
              <button style={{ ...s.btn, ...s.btnGold, flex: 1 }}
                onClick={() => { createGame.mutate(form); setShowModal(false) }}>
                Create Table
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
