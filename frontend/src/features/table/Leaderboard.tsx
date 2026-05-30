import { useTranslation } from 'react-i18next'
import type { LeaderboardEntry } from './useTable'

interface LeaderboardProps {
  entries: LeaderboardEntry[]
  dealerUserId: string
  currentUserId: string | undefined
  canKick: boolean
  onKick: (playerId: string) => void
}

export function Leaderboard({
  entries, dealerUserId, currentUserId, canKick, onKick,
}: LeaderboardProps) {
  const { t } = useTranslation(['common', 'table'])
  return (
    <div>
      <h3 style={{ color: '#e2c97e', marginBottom: '0.75rem' }}>{t('table:leaderboard.title')}</h3>
      {[...entries].sort((a, b) => a.seatOrder - b.seatOrder).map((e) => (
        <div key={e.playerId} style={{
          display: 'flex', justifyContent: 'space-between', alignItems: 'center',
          padding: '0.5rem 0.75rem', borderRadius: '8px', marginBottom: '0.4rem',
          background: '#1a2a40', border: '1px solid #2d4a6a',
        }}>
          <span style={{ color: '#e2e8f0' }}>
            {e.username}
            {e.userId === dealerUserId ? ` ${t('dealer')}` : ''}
            {e.userId === currentUserId ? ` ${t('you')}` : ''}
          </span>
          {canKick && e.userId !== currentUserId && (
            <button
              onClick={() => onKick(e.playerId)}
              style={{ padding: '0.2rem 0.5rem', background: 'transparent', border: '1px solid #f87171', borderRadius: '4px', color: '#f87171', cursor: 'pointer', fontSize: '0.75rem' }}
            >
              {t('table:leaderboard.kick')}
            </button>
          )}
        </div>
      ))}
    </div>
  )
}
