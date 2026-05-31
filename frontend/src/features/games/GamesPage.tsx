import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useRequireAuth } from '../../shared/hooks/useRequireAuth'
import { useAuthStore } from '../auth/authStore'
import { useGames, useCreateGame, useJoinGame, useDeleteGame } from './useGames'
import { LocaleSwitcher } from '../../shared/components/LocaleSwitcher'
import apiClient from '../../lib/apiClient'

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
  const { t } = useTranslation(['common', 'lobby', 'table', 'errors'])
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const navigate = useNavigate()
  const { data: games, isLoading } = useGames()
  const createGame = useCreateGame()
  const joinGame = useJoinGame()
  const deleteGame = useDeleteGame()
  const isAdmin = user?.role === 'admin'
  const [showModal, setShowModal] = useState(false)
  const [joinErrorGameId, setJoinErrorGameId] = useState<string | null>(null)

  const handleJoin = (gameId: string) => {
    setJoinErrorGameId(null)
    joinGame.mutate(gameId, {
      onError: () => setJoinErrorGameId(gameId),
    })
  }

  const openGames = games?.filter(g => g.status !== 'FINISHED') ?? []

  const myOrphanedTable = openGames.find(g => g.dealerUserId === user?.id)
  const [form, setForm] = useState({ deckCount: 2, minPlayers: 2, maxPlayers: 8 })

  if (!authed) return null

  return (
    <div style={s.page}>
      <div style={s.header}>
        <span style={s.title}>{t('appName')}</span>
        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          <span style={{ color: '#7a9bb5' }}>👤 {user?.username}</span>
          <LocaleSwitcher />
          <button style={{ ...s.btn, ...s.btnOutline }} onClick={() => { logout(); navigate('/login') }}>{t('logout')}</button>
        </div>
      </div>

      {myOrphanedTable && (
        <div style={{ background: '#1a2a1a', border: '1px solid #4ade8088', borderRadius: '10px', padding: '1rem 1.25rem', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '1rem' }}>
          <span style={{ color: '#4ade80' }}>⚠ {t('lobby:orphanWarning', { status: t(`table:status.${myOrphanedTable.status}`) })}</span>
          <div style={{ display: 'flex', gap: '0.5rem', flexShrink: 0 }}>
            <button style={{ ...s.btn, ...s.btnGold }} onClick={() => navigate(`/table/${myOrphanedTable.id}`)}>{t('lobby:return')}</button>
            <button style={{ ...s.btn, padding: '0.5rem 0.9rem', background: 'transparent', color: '#f87171', border: '1px solid #f87171' }} onClick={() => deleteGame.mutate(myOrphanedTable.id)}>{t('lobby:delete')}</button>
          </div>
        </div>
      )}

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h2 style={{ color: '#7a9bb5', fontWeight: 400, margin: 0 }}>{t('lobby:openTables')}</h2>
        <button
          disabled={!!myOrphanedTable}
          title={myOrphanedTable ? t('lobby:newTableBlockedTitle') : undefined}
          style={{ ...s.btn, ...s.btnGold, opacity: myOrphanedTable ? 0.4 : 1, cursor: myOrphanedTable ? 'not-allowed' : 'pointer' }}
          onClick={() => { if (!myOrphanedTable) { createGame.reset(); setShowModal(true) } }}
        >{t('lobby:newTable')}</button>
      </div>

      {isLoading && <p style={{ color: '#4a6a8a' }}>{t('lobby:loadingTables')}</p>}
      {!isLoading && openGames.length === 0 && (
        <p style={{ color: '#4a6a8a' }}>{t('lobby:noTables')}</p>
      )}

      <div style={s.grid}>
        {openGames.map(game => (
          <div key={game.id} style={s.card}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
              <span style={{ color: '#e2c97e', fontWeight: 600 }}>{t('lobby:dealerLabel', { name: game.dealerUsername })}</span>
              <span style={{ ...s.badge, ...statusColor(game.status) }}>{t(`table:status.${game.status}`)}</span>
            </div>
            <div style={{ color: '#4a6a8a', fontSize: '0.75rem', marginBottom: '0.75rem' }}>
              {new Date(game.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </div>
            <div style={{ color: '#7a9bb5', fontSize: '0.85rem', lineHeight: 1.8 }}>
              <div>🃏 {t('lobby:decks', { count: game.deckCount, total: game.deckCount * 52 })}</div>
              <div>👥 {t('lobby:players', { current: game.playerCount, max: game.maxPlayers })}</div>
            </div>
            {joinErrorGameId === game.id && (
              <p style={{ color: '#f87171', fontSize: '0.8rem', margin: '0.5rem 0 0' }}>
                {t(`errors:${(joinGame.error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'UNKNOWN'}`,
                  { defaultValue: t('errors:UNKNOWN') })}
              </p>
            )}
            <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.5rem' }}>
              {(() => {
                const shoeTooLow = game.status === 'IN_PROGRESS' && game.remainingCards === 0
                const disabled = !!myOrphanedTable || shoeTooLow
                return (
                  <button
                    disabled={disabled}
                    title={shoeTooLow ? t('errors:SHOE_TOO_LOW_FOR_JOIN') : undefined}
                    style={{ ...s.btn, ...s.btnGold, flex: 1, opacity: disabled ? 0.4 : 1, cursor: disabled ? 'not-allowed' : 'pointer' }}
                    onClick={() => !disabled && handleJoin(game.id)}
                  >
                    {t('lobby:joinTable')}
                  </button>
                )
              })()}
              {isAdmin && (
                <button
                  style={{ ...s.btn, background: 'transparent', color: '#f87171', border: '1px solid #f87171', padding: '0.5rem 0.75rem' }}
                  onClick={() => deleteGame.mutate(game.id)}
                  title={t('lobby:adminDelete')}
                >
                  🗑
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      {import.meta.env.DEV && <DebugPanel />}

      {showModal && (
        <div style={s.modal}>
          <div style={s.modalBox}>
            <h2 style={{ color: '#e2c97e', marginBottom: '1rem' }}>{t('lobby:modal.title')}</h2>
            <label style={s.label}>{t('lobby:modal.deckCount')}</label>
            <input type="number" min={1} max={8} value={form.deckCount} style={s.input}
              onChange={e => setForm({ ...form, deckCount: +e.target.value })} />
            <label style={s.label}>{t('lobby:modal.minPlayers')}</label>
            <input type="number" min={2} max={form.maxPlayers} value={form.minPlayers} style={s.input}
              onChange={e => setForm({ ...form, minPlayers: +e.target.value })} />
            <label style={s.label}>{t('lobby:modal.maxPlayers')}</label>
            <input type="number" min={form.minPlayers} max={8} value={form.maxPlayers} style={s.input}
              onChange={e => setForm({ ...form, maxPlayers: +e.target.value })} />
            {createGame.isError && (
              <p style={{ color: '#f87171', fontSize: '0.8rem', margin: '0.75rem 0 0' }}>
                {t(`errors:${(createGame.error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'UNKNOWN'}`,
                  { defaultValue: t('errors:UNKNOWN') })}
              </p>
            )}
            <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.5rem' }}>
              <button style={{ ...s.btn, ...s.btnOutline, flex: 1 }} onClick={() => { createGame.reset(); setShowModal(false) }}>{t('lobby:modal.cancel')}</button>
              <button style={{ ...s.btn, ...s.btnGold, flex: 1 }}
                onClick={() => createGame.mutate(form, { onSuccess: () => setShowModal(false) })}>
                {t('lobby:modal.create')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function DebugPanel() {
  const [beStatus, setBeStatus] = useState<string | null>(null)

  const trigger = async (path: string) => {
    try {
      const res = await apiClient.get(path)
      setBeStatus(`✓ ${path} → ${res.status} (check server logs)`)
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status ?? '?'
      setBeStatus(`⚠ ${path} → ${status} (check server logs)`)
    }
  }

  return (
    <div style={{ marginTop: '3rem', borderTop: '1px solid #1a2a40', paddingTop: '1rem', textAlign: 'center' }}>
      <span style={{ color: '#4a6a8a', fontSize: '0.72rem', marginRight: '0.75rem' }}>🔧 Debug / Observability:</span>
      <button
        onClick={() => trigger('/debug/error')}
        style={{ color: '#f87171', background: 'none', border: 'none', cursor: 'pointer', fontSize: '0.72rem', marginRight: '0.75rem', textDecoration: 'underline' }}
      >
        Trigger BE Error (500)
      </button>
      <button
        onClick={() => trigger('/debug/warn')}
        style={{ color: '#fbbf24', background: 'none', border: 'none', cursor: 'pointer', fontSize: '0.72rem', textDecoration: 'underline' }}
      >
        Trigger BE Warning
      </button>
      {beStatus && (
        <p style={{ color: '#4a6a8a', fontSize: '0.72rem', margin: '0.4rem 0 0' }}>{beStatus}</p>
      )}
    </div>
  )
}
