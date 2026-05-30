import { CardBadge } from './CardBadge'
import type { LeaderboardEntry, ShoeCard } from './useTable'

interface GameResultProps {
  leaderboard: LeaderboardEntry[]
  hand: ShoeCard[] | undefined
  dealerUserId: string
  currentUserId: string | undefined
  myHandValue: number
  onLobby: () => void
}

export function GameResult({
  leaderboard, hand, dealerUserId, currentUserId, myHandValue, onLobby,
}: GameResultProps) {
  const topScore = leaderboard[0]?.handValue ?? 0
  const topPlayers = leaderboard.filter(e => e.handValue === topScore)
  const isDraw = topPlayers.length > 1

  return (
    <div style={{ flex: 1, background: '#0f1a2e', color: '#e2e8f0', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '2rem' }}>
      <div style={{ background: '#1a2a40', borderRadius: '16px', padding: '2.5rem', width: '100%', maxWidth: '480px', border: '1px solid #e2c97e44', textAlign: 'center' }}>
        <div style={{ fontSize: '3rem', marginBottom: '0.5rem' }}>{isDraw ? '🤝' : '🏆'}</div>
        <h1 style={{ color: '#e2c97e', margin: '0 0 0.25rem' }}>Game Over</h1>
        {isDraw ? (
          <p style={{ color: '#60a5fa', marginBottom: '1.5rem' }}>
            Draw — <strong>{topPlayers.map(p => p.username).join(', ')}</strong> — {topScore} pts
          </p>
        ) : (
          <p style={{ color: '#4ade80', marginBottom: '1.5rem' }}>
            Winner: <strong>{topPlayers[0]?.username}</strong> — {topScore} pts
          </p>
        )}
        <div style={{ textAlign: 'left', marginBottom: '1.5rem' }}>
          <h3 style={{ color: '#7a9bb5', marginBottom: '0.75rem', fontWeight: 400 }}>Final Standings</h3>
          {leaderboard.map((e, i) => {
            const isTop = e.handValue === topScore
            return (
              <div key={e.playerId} style={{
                display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                padding: '0.5rem 0.75rem', borderRadius: '8px', marginBottom: '0.4rem',
                background: isTop ? '#1a3a20' : '#0f1a2e', border: '1px solid #2d4a6a',
              }}>
                <span style={{ color: isTop ? '#4ade80' : '#e2e8f0' }}>
                  #{i + 1} {e.username}{e.userId === dealerUserId ? ' 🎩' : ''}
                  {e.userId === currentUserId ? ' (you)' : ''}
                </span>
                <span style={{ color: '#e2c97e', fontWeight: 700 }}>{e.handValue} pts · {e.cardCount} cards</span>
              </div>
            )
          })}
        </div>
        {hand && hand.length > 0 && (
          <div style={{ marginBottom: '1.5rem', textAlign: 'left' }}>
            <h3 style={{ color: '#7a9bb5', marginBottom: '0.5rem', fontWeight: 400 }}>Your final hand ({myHandValue} pts)</h3>
            <div style={{ display: 'flex', flexWrap: 'wrap' }}>
              {hand.map(c => <CardBadge key={c.id} suit={c.suit} face={c.face} />)}
            </div>
          </div>
        )}
        <button
          onClick={onLobby}
          style={{ padding: '0.75rem 2rem', background: '#e2c97e', color: '#0f1a2e', border: 'none', borderRadius: '8px', fontWeight: 700, fontSize: '1rem', cursor: 'pointer', width: '100%' }}
        >
          Return to Lobby
        </button>
      </div>
    </div>
  )
}
