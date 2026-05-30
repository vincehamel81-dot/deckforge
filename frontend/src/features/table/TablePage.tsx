import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useRequireAuth } from '../../shared/hooks/useRequireAuth'
import { useAuthStore } from '../auth/authStore'
import {
  useGameDetail, useLeaderboard, usePlayerHand, useSuitCounts,
  useStartGame, useEndGame, useShuffle, useDealToAll, useAddDeck, useLeaveGame,
  type LeaderboardEntry,
} from './useTable'

const SUIT_SYMBOL: Record<string, string> = {
  HEARTS: '♥', SPADES: '♠', CLUBS: '♣', DIAMONDS: '♦',
}
const SUIT_COLOR: Record<string, string> = {
  HEARTS: '#f87171', SPADES: '#e2e8f0', CLUBS: '#e2e8f0', DIAMONDS: '#f87171',
}

function CardBadge({ suit, face }: { suit: string; face: string }) {
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
      background: '#0f1a2e', border: '1px solid #2d4a6a', borderRadius: '6px',
      padding: '0.3rem 0.5rem', margin: '0.2rem', fontSize: '0.9rem', fontWeight: 700,
      color: SUIT_COLOR[suit] ?? '#e2e8f0', minWidth: '2.8rem',
    }}>
      {face.slice(0, 1) === 'T' ? '10' : face[0]}{SUIT_SYMBOL[suit]}
    </span>
  )
}

function Leaderboard({ entries, dealerUserId }: { entries: LeaderboardEntry[]; dealerUserId: string }) {
  return (
    <div>
      <h3 style={{ color: '#e2c97e', marginBottom: '0.5rem' }}>🏆 Leaderboard</h3>
      {entries.map((e, i) => (
        <div key={e.playerId} style={{
          display: 'flex', justifyContent: 'space-between', alignItems: 'center',
          padding: '0.5rem 0.75rem', borderRadius: '8px', marginBottom: '0.4rem',
          background: i === 0 ? '#1a3a20' : '#1a2a40', border: '1px solid #2d4a6a',
        }}>
          <span style={{ color: i === 0 ? '#4ade80' : '#e2e8f0' }}>
            #{i + 1} {e.username}{e.userId === dealerUserId ? ' 🎩' : ''}
          </span>
          <span style={{ color: '#e2c97e', fontWeight: 700 }}>{e.handValue} pts ({e.cardCount} cards)</span>
        </div>
      ))}
    </div>
  )
}

export default function TablePage() {
  const authed = useRequireAuth()
  const { id: gameId } = useParams<{ id: string }>()
  const user = useAuthStore(s => s.user)
  const navigate = useNavigate()

  const { data: detail, isLoading } = useGameDetail(gameId!)
  const { data: leaderboard } = useLeaderboard(gameId!)
  const { data: suitCounts } = useSuitCounts(gameId!)

  // Find current user's player entry from leaderboard
  const myEntry = leaderboard?.find(e => e.userId === user?.id)
  const { data: hand } = usePlayerHand(gameId!, myEntry?.playerId)

  const startGame = useStartGame(gameId!)
  const endGame = useEndGame(gameId!)
  const shuffle = useShuffle(gameId!)
  const dealToAll = useDealToAll(gameId!)
  const addDeck = useAddDeck(gameId!)
  const leaveGame = useLeaveGame(gameId!)

  const [dealCount, setDealCount] = useState(1)
  const [startDeal, setStartDeal] = useState(2)
  const [shuffleMsg, setShuffleMsg] = useState('')

  if (!authed) return null
  if (isLoading || !detail) return (
    <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#7a9bb5', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      Loading table...
    </div>
  )

  const { game, totalCards, remainingCards } = detail
  const isDealer = user?.id === game.dealerUserId
  const drawsRemaining = leaderboard && leaderboard.length > 0
    ? Math.floor(remainingCards / leaderboard.length)
    : 0

  return (
    <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#e2e8f0', padding: '1.5rem' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <span style={{ color: '#e2c97e', fontWeight: 700, fontSize: '1.3rem' }}>♠ DeckForge</span>
          <span style={{ color: '#4a6a8a', marginLeft: '1rem', fontSize: '0.85rem' }}>Table {gameId?.slice(0, 8)}…</span>
        </div>
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <span style={{
            padding: '0.2rem 0.7rem', borderRadius: '4px', fontSize: '0.8rem', fontWeight: 600,
            background: game.status === 'WAITING' ? '#1a4a2a' : game.status === 'IN_PROGRESS' ? '#1a2a4a' : '#2a1a1a',
            color: game.status === 'WAITING' ? '#4ade80' : game.status === 'IN_PROGRESS' ? '#60a5fa' : '#f87171',
          }}>
            {game.status}
          </span>
          <button
            onClick={async () => {
              if (myEntry && game.status !== 'FINISHED') {
                if (!window.confirm('Leave this table? Your cards will be returned to the shoe.')) return
                await leaveGame.mutateAsync(myEntry.playerId)
              }
              navigate('/')
            }}
            style={{ padding: '0.4rem 0.8rem', background: 'transparent', border: '1px solid #4a6a8a', borderRadius: '6px', color: '#7a9bb5', cursor: 'pointer' }}
          >← Lobby</button>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 300px', gap: '1.5rem' }}>
        {/* Main area */}
        <div>
          {/* Shoe status */}
          <div style={{ background: '#1a2a40', borderRadius: '12px', padding: '1rem', marginBottom: '1rem', border: '1px solid #2d4a6a' }}>
            <h3 style={{ color: '#e2c97e', marginBottom: '0.75rem' }}>🃏 Shoe Status</h3>
            <div style={{ display: 'flex', gap: '2rem', flexWrap: 'wrap' }}>
              <div><span style={{ color: '#4a6a8a' }}>Total: </span><strong>{totalCards}</strong></div>
              <div><span style={{ color: '#4a6a8a' }}>Remaining: </span><strong style={{ color: remainingCards < 10 ? '#f87171' : '#4ade80' }}>{remainingCards}</strong></div>
              <div><span style={{ color: '#4a6a8a' }}>Draws left: </span><strong style={{ color: '#e2c97e' }}>{drawsRemaining}</strong></div>
              <div><span style={{ color: '#4a6a8a' }}>Decks: </span><strong>{game.deckCount}</strong></div>
            </div>
            {suitCounts && (
              <div style={{ display: 'flex', gap: '1rem', marginTop: '0.75rem', flexWrap: 'wrap' }}>
                {(suitCounts as { suit: string; count: number }[]).map(s => (
                  <span key={s.suit} style={{ color: SUIT_COLOR[s.suit], fontSize: '0.9rem' }}>
                    {SUIT_SYMBOL[s.suit]} {s.count}
                  </span>
                ))}
              </div>
            )}
          </div>

          {/* Your hand */}
          <div style={{ background: '#1a2a40', borderRadius: '12px', padding: '1rem', marginBottom: '1rem', border: '1px solid #2d4a6a' }}>
            <h3 style={{ color: '#e2c97e', marginBottom: '0.75rem' }}>
              🤲 Your Hand {myEntry && <span style={{ color: '#7a9bb5', fontWeight: 400 }}>({myEntry.handValue} pts)</span>}
            </h3>
            {!hand || hand.length === 0
              ? <p style={{ color: '#4a6a8a' }}>No cards yet</p>
              : <div style={{ display: 'flex', flexWrap: 'wrap' }}>{hand.map(c => <CardBadge key={c.id} suit={c.suit} face={c.face} />)}</div>
            }
          </div>

          {/* Dealer controls */}
          {isDealer && (
            <div style={{ background: '#1a2a40', borderRadius: '12px', padding: '1rem', border: '1px solid #e2c97e44' }}>
              <h3 style={{ color: '#e2c97e', marginBottom: '1rem' }}>🎩 Dealer Controls</h3>
              <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', alignItems: 'center' }}>
                {game.status === 'WAITING' && (
                  <>
                    <button onClick={() => addDeck.mutate()} style={{ padding: '0.5rem 1rem', background: '#1a3a4a', color: '#60a5fa', border: '1px solid #60a5fa', borderRadius: '8px', cursor: 'pointer' }}>
                      + Add Deck
                    </button>
                    <button
                      onClick={async () => {
                        await shuffle.mutateAsync()
                        setShuffleMsg('✓ Shuffled!')
                        setTimeout(() => setShuffleMsg(''), 2500)
                      }}
                      style={{ padding: '0.5rem 1rem', background: '#1a2a4a', color: '#a78bfa', border: '1px solid #a78bfa', borderRadius: '8px', cursor: 'pointer' }}
                    >
                      🔀 Shuffle
                    </button>
                    {shuffleMsg && <span style={{ color: '#a78bfa', fontSize: '0.85rem' }}>{shuffleMsg}</span>}
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <span style={{ color: '#7a9bb5', fontSize: '0.85rem' }}>Initial deal:</span>
                      <input type="number" min={0} max={10} value={startDeal} onChange={e => setStartDeal(+e.target.value)}
                        style={{ width: '50px', padding: '0.4rem', background: '#0f1a2e', border: '1px solid #2d4a6a', borderRadius: '6px', color: '#e2e8f0', textAlign: 'center' }} />
                    </div>
                    {(leaderboard?.length ?? 0) < game.minPlayers ? (
                      <span style={{ color: '#7a9bb5', fontSize: '0.85rem' }}>
                        ⏳ Waiting for players ({leaderboard?.length ?? 0}/{game.minPlayers})
                      </span>
                    ) : (
                      <button onClick={() => startGame.mutate(startDeal)} style={{ padding: '0.5rem 1.2rem', background: '#1a4a2a', color: '#4ade80', border: '1px solid #4ade80', borderRadius: '8px', cursor: 'pointer', fontWeight: 700 }}>
                        ▶ Start Game
                      </button>
                    )}
                    {startGame.isError && (
                      <span style={{ color: '#f87171', fontSize: '0.85rem' }}>
                        {(startGame.error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Could not start game'}
                      </span>
                    )}
                  </>
                )}
                {game.status === 'IN_PROGRESS' && (
                  <>
                    <button
                      onClick={async () => {
                        await shuffle.mutateAsync()
                        setShuffleMsg('✓ Shuffled!')
                        setTimeout(() => setShuffleMsg(''), 2500)
                      }}
                      style={{ padding: '0.5rem 1rem', background: '#1a2a4a', color: '#a78bfa', border: '1px solid #a78bfa', borderRadius: '8px', cursor: 'pointer' }}
                    >
                      🔀 Shuffle
                    </button>
                    {shuffleMsg && <span style={{ color: '#a78bfa', fontSize: '0.85rem' }}>{shuffleMsg}</span>}
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <span style={{ color: '#7a9bb5', fontSize: '0.85rem' }}>Deal cards:</span>
                      <input type="number" min={1} max={10} value={dealCount} onChange={e => setDealCount(+e.target.value)}
                        style={{ width: '50px', padding: '0.4rem', background: '#0f1a2e', border: '1px solid #2d4a6a', borderRadius: '6px', color: '#e2e8f0', textAlign: 'center' }} />
                    </div>
                    <button
                      onClick={() => dealToAll.mutate({ playerIds: (leaderboard ?? []).map(e => e.playerId), count: dealCount })}
                      style={{ padding: '0.5rem 1rem', background: '#e2c97e', color: '#0f1a2e', border: 'none', borderRadius: '8px', cursor: 'pointer', fontWeight: 700 }}>
                      🃏 Deal to All
                    </button>
                    <button onClick={() => endGame.mutate()} style={{ padding: '0.5rem 1rem', background: '#2a1a1a', color: '#f87171', border: '1px solid #f87171', borderRadius: '8px', cursor: 'pointer' }}>
                      ⏹ End Game
                    </button>
                  </>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Sidebar: leaderboard */}
        <div style={{ background: '#1a2a40', borderRadius: '12px', padding: '1rem', border: '1px solid #2d4a6a', alignSelf: 'start' }}>
          {leaderboard && leaderboard.length > 0
            ? <Leaderboard entries={leaderboard} dealerUserId={game.dealerUserId} />
            : <p style={{ color: '#4a6a8a' }}>No players yet</p>
          }
        </div>
      </div>
    </div>
  )
}
