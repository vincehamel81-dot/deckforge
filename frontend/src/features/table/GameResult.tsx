import { useTranslation } from 'react-i18next'
import { CardBadge } from './CardBadge'
import { usePlayerHand } from './useTable'
import type { LeaderboardEntry } from './useTable'

interface GameResultProps {
  leaderboard: LeaderboardEntry[]
  dealerUserId: string
  currentUserId: string | undefined
  gameId: string
  onLobby: () => void
}

function PlayerHand({ gameId, playerId }: { gameId: string; playerId: string }) {
  const { data: hand } = usePlayerHand(gameId, playerId)
  if (!hand || hand.length === 0) return null
  return (
    <div style={{ display: 'flex', flexWrap: 'wrap', marginTop: '0.4rem' }}>
      {hand.map(c => <CardBadge key={c.id} suit={c.suit} face={c.face} />)}
    </div>
  )
}

export function GameResult({
  leaderboard, dealerUserId, currentUserId, gameId, onLobby,
}: GameResultProps) {
  const { t } = useTranslation(['common', 'table'])
  const topScore = leaderboard[0]?.handValue ?? 0
  const topPlayers = leaderboard.filter(e => e.handValue === topScore)
  const isDraw = topPlayers.length > 1

  return (
    <div style={{ flex: 1, background: '#0f1a2e', color: '#e2e8f0', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '2rem' }}>
      <div style={{ background: '#1a2a40', borderRadius: '16px', padding: '2.5rem', width: '100%', maxWidth: '520px', border: '1px solid #e2c97e44', textAlign: 'center' }}>
        <div style={{ fontSize: '3rem', marginBottom: '0.75rem' }}>{isDraw ? '🤝' : '🏆'}</div>
        <h1 style={{ color: '#e2c97e', margin: '0 0 0.75rem', fontSize: '1.6rem' }}>{t('table:result.gameOver')}</h1>
        {isDraw ? (
          <p style={{ color: '#60a5fa', marginBottom: '1.5rem', overflowWrap: 'break-word' }}>
            {t('table:result.draw', { players: topPlayers.map(p => p.username).join(', '), score: topScore })}
          </p>
        ) : (
          <p style={{ color: '#4ade80', marginBottom: '1.5rem', overflowWrap: 'break-word' }}>
            {t('table:result.winner', { player: topPlayers[0]?.username, score: topScore })}
          </p>
        )}

        <div style={{ textAlign: 'left', marginBottom: '1.5rem' }}>
          <h3 style={{ color: '#7a9bb5', marginBottom: '0.75rem', fontWeight: 400 }}>{t('table:result.finalStandings')}</h3>
          {leaderboard.map((e, i) => {
            const isTop = e.handValue === topScore
            return (
              <div key={e.playerId} style={{
                padding: '0.6rem 0.75rem', borderRadius: '8px', marginBottom: '0.5rem',
                background: isTop ? '#1a3a20' : '#0f1a2e', border: '1px solid #2d4a6a',
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ color: isTop ? '#4ade80' : '#e2e8f0' }}>
                    #{i + 1} {e.username}
                    {e.userId === dealerUserId ? ` ${t('dealer')}` : ''}
                    {e.userId === currentUserId ? ` ${t('you')}` : ''}
                  </span>
                  <span style={{ color: '#e2c97e', fontWeight: 700 }}>
                    {t('table:result.ptsCards', { pts: e.handValue, cards: e.cardCount })}
                  </span>
                </div>
                <PlayerHand gameId={gameId} playerId={e.playerId} />
              </div>
            )
          })}
        </div>

        <button
          onClick={onLobby}
          style={{ padding: '0.75rem 2rem', background: '#e2c97e', color: '#0f1a2e', border: 'none', borderRadius: '8px', fontWeight: 700, fontSize: '1rem', cursor: 'pointer', width: '100%' }}
        >
          {t('returnToLobby')}
        </button>
      </div>
    </div>
  )
}
