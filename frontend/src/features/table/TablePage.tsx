import { useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import { useRequireAuth } from '../../shared/hooks/useRequireAuth'
import { useAuthStore } from '../auth/authStore'
import {
  useGameDetail, useLeaderboard, usePlayerHand, useSuitCounts, useCardCounts,
  useStartGame, useEndGame, useShuffle, useDealToAll, useAddDeck, useLeaveGame, useDeleteGame,
} from './useTable'
import { useGameSocket } from './useGameSocket'
import { CardBadge, SUIT_SYMBOL, SUIT_COLOR, FACE_LABEL } from './CardBadge'
import { Leaderboard } from './Leaderboard'
import { GameResult } from './GameResult'
import { DealerControls } from './DealerControls'

export default function TablePage() {
  const authed = useRequireAuth()
  const { id: gameId } = useParams<{ id: string }>()
  const user = useAuthStore(s => s.user)
  const navigate = useNavigate()

  const [tableClosed, setTableClosed] = useState(false)
  useGameSocket(gameId!, { onGameDeleted: () => setTableClosed(true) })

  const { data: detail, isLoading, isError } = useGameDetail(gameId!)
  const { data: leaderboard } = useLeaderboard(gameId!)
  const myEntry = leaderboard?.find(e => e.userId === user?.id)
  const { data: hand } = usePlayerHand(gameId!, myEntry?.playerId)
  const { data: suitCounts } = useSuitCounts(gameId!)
  const qc = useQueryClient()
  const [showShoeDetails, setShowShoeDetails] = useState(false)
  const { data: cardCounts } = useCardCounts(gameId!)

  const startGame = useStartGame(gameId!)
  const endGame = useEndGame(gameId!)
  const shuffle = useShuffle(gameId!)
  const dealToAll = useDealToAll(gameId!)
  const addDeck = useAddDeck(gameId!)
  const leaveGame = useLeaveGame(gameId!)
  const deleteGame = useDeleteGame(gameId!)

  if (!authed) return null

  if (tableClosed || isError) return (
    <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#e2e8f0', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <div style={{ background: '#1a2a40', borderRadius: '16px', padding: '2.5rem', maxWidth: '400px', width: '100%', textAlign: 'center', border: '1px solid #f8717144' }}>
        <div style={{ fontSize: '2.5rem', marginBottom: '0.5rem' }}>🚪</div>
        <h2 style={{ color: '#f87171', margin: '0 0 0.75rem' }}>Table Closed</h2>
        <p style={{ color: '#7a9bb5', marginBottom: '1.5rem' }}>The dealer closed this table. All cards have been returned to the shoe.</p>
        <button
          onClick={() => navigate('/', { replace: true })}
          style={{ padding: '0.75rem 2rem', background: '#e2c97e', color: '#0f1a2e', border: 'none', borderRadius: '8px', fontWeight: 700, cursor: 'pointer', width: '100%' }}
        >
          Return to Lobby
        </button>
      </div>
    </div>
  )

  if (isLoading || !detail) return (
    <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#7a9bb5', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      Loading table...
    </div>
  )

  const { game, totalCards, remainingCards, dealerUsername } = detail
  const isDealer = user?.id === game.dealerUserId
  const isAdmin = user?.role === 'admin'
  const drawsRemaining = leaderboard && leaderboard.length > 0
    ? Math.floor(remainingCards / leaderboard.length)
    : 0
  const dealerEntry = leaderboard?.find(e => e.userId === game.dealerUserId)
  const displayDealerName = dealerUsername || dealerEntry?.username || `${game.dealerUserId.slice(0, 8)}…`

  if (game.status === 'FINISHED') {
    if (!leaderboard) return (
      <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#7a9bb5', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        Loading results…
      </div>
    )
    return (
      <GameResult
        leaderboard={leaderboard}
        hand={hand}
        dealerUserId={game.dealerUserId}
        currentUserId={user?.id}
        myHandValue={myEntry?.handValue ?? 0}
        onLobby={() => navigate('/')}
      />
    )
  }

  const startGameError = startGame.isError
    ? (startGame.error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Could not start game'
    : undefined

  return (
    <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#e2e8f0', padding: '1.5rem' }}>

      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <span style={{ color: '#e2c97e', fontWeight: 700, fontSize: '1.3rem' }}>♠ DeckForge</span>
          <span style={{ color: '#4a6a8a', marginLeft: '1rem', fontSize: '0.85rem' }}>
            Table created by <span style={{ color: '#e2c97e' }}>{displayDealerName}</span>
          </span>
        </div>
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <span style={{
            padding: '0.2rem 0.7rem', borderRadius: '4px', fontSize: '0.8rem', fontWeight: 600,
            background: game.status === 'WAITING' ? '#1a4a2a' : '#1a2a4a',
            color: game.status === 'WAITING' ? '#4ade80' : '#60a5fa',
          }}>
            {game.status}
          </span>
          <button
            onClick={async () => {
              if (myEntry && game.status !== 'FINISHED') {
                if (isDealer) {
                  if (!window.confirm('You are the dealer. Leaving will close this table for everyone and delete it. Proceed?')) return
                  await deleteGame.mutateAsync()
                  navigate('/')
                  return
                }
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

        {/* Main column */}
        <div>
          {/* Shoe status */}
          <div style={{ background: '#1a2a40', borderRadius: '12px', padding: '1rem', marginBottom: '1rem', border: '1px solid #2d4a6a' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
              <h3 style={{ color: '#e2c97e', margin: 0 }}>🃏 Shoe Status</h3>
              <button
                onClick={() => {
                  if (!showShoeDetails) {
                    // Explicit user-initiated API call — force a fresh fetch, not cached data.
                    qc.refetchQueries({ queryKey: ['suits', gameId!] })
                  }
                  setShowShoeDetails(v => !v)
                }}
                style={{ padding: '0.2rem 0.6rem', background: 'transparent', border: '1px solid #4a6a8a', borderRadius: '4px', color: '#7a9bb5', cursor: 'pointer', fontSize: '0.75rem' }}
              >
                {showShoeDetails ? 'Hide Suits' : 'Check Suits'}
              </button>
            </div>
            <div style={{ display: 'flex', gap: '2rem', flexWrap: 'wrap', marginBottom: '0.75rem' }}>
              <div><span style={{ color: '#4a6a8a' }}>Total: </span><strong>{totalCards}</strong></div>
              <div><span style={{ color: '#4a6a8a' }}>Remaining: </span><strong style={{ color: remainingCards < 10 ? '#f87171' : '#4ade80' }}>{remainingCards}</strong></div>
              <div><span style={{ color: '#4a6a8a' }}>Draws left: </span><strong style={{ color: '#e2c97e' }}>{drawsRemaining}</strong></div>
              <div><span style={{ color: '#4a6a8a' }}>Decks: </span><strong>{game.deckCount}</strong></div>
            </div>
            {suitCounts && (
              <div style={{ display: 'flex', gap: '1.5rem', flexWrap: 'wrap' }}>
                {suitCounts.map(s => (
                  <span key={s.suit} style={{ fontSize: '0.85rem' }}>
                    <span style={{ color: SUIT_COLOR[s.suit] ?? '#e2e8f0' }}>
                      {SUIT_SYMBOL[s.suit] ?? '?'}
                    </span>
                    {' '}<strong style={{ color: '#e2e8f0' }}>{s.count}</strong>
                  </span>
                ))}
              </div>
            )}
            {showShoeDetails && suitCounts && (
              <div style={{ marginTop: '0.75rem', borderTop: '1px solid #2d4a6a', paddingTop: '0.75rem' }}>
                <p style={{ color: '#4a6a8a', fontSize: '0.75rem', margin: '0 0 0.5rem' }}>Undealt cards per suit</p>
                <div style={{ display: 'flex', gap: '1.5rem', flexWrap: 'wrap' }}>
                  {suitCounts.map(s => (
                    <span key={s.suit} style={{ fontSize: '0.85rem' }}>
                      <span style={{ color: SUIT_COLOR[s.suit] ?? '#e2e8f0' }}>
                        {s.suit === 'HEARTS' ? 'Hearts' : s.suit === 'SPADES' ? 'Spades' : s.suit === 'CLUBS' ? 'Clubs' : 'Diamonds'}
                      </span>
                      <span style={{ color: '#e2e8f0' }}>: <strong>{s.count}</strong></span>
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* A7: warn when shoe can't serve a full round */}
          {game.status === 'IN_PROGRESS' && leaderboard && leaderboard.length > 0 && remainingCards < leaderboard.length && (
            <div style={{ background: '#2a1a1a', border: '1px solid #f8717166', borderRadius: '8px', padding: '0.6rem 0.9rem', marginBottom: '1rem', fontSize: '0.85rem', color: '#f87171' }}>
              {remainingCards === 0
                ? '⚠ Shoe exhausted — no cards remaining. End the game to see final standings.'
                : `⚠ Only ${remainingCards} card${remainingCards === 1 ? '' : 's'} left — not enough for a full round. End the game or deal partial hands.`}
            </div>
          )}

          {/* Player hand */}
          <div style={{ background: '#1a2a40', borderRadius: '12px', padding: '1rem', marginBottom: '1rem', border: '1px solid #2d4a6a' }}>
            <h3 style={{ color: '#e2c97e', marginBottom: '0.75rem' }}>
              🤲 Your Hand {myEntry && <span style={{ color: '#7a9bb5', fontWeight: 400 }}>({myEntry.handValue} pts)</span>}
            </h3>
            {!hand || hand.length === 0
              ? <p style={{ color: '#4a6a8a' }}>No cards yet</p>
              : <div style={{ display: 'flex', flexWrap: 'wrap' }}>
                  {hand.map((c, i) => (
                    <CardBadge key={c.id} suit={c.suit} face={c.face} isNew={i === hand.length - 1} />
                  ))}
                </div>
            }
          </div>

          {(game.status === 'WAITING' || game.status === 'IN_PROGRESS') && isDealer && (
            <DealerControls
              gameStatus={game.status}
              playerCount={leaderboard?.length ?? 0}
              minPlayers={game.minPlayers}
              onAddDeck={() => addDeck.mutate()}
              onShuffle={async () => { await shuffle.mutateAsync() }}
              onStartGame={(n) => startGame.mutate(n)}
              onDealToAll={(n) => dealToAll.mutate({ count: n })}
              onEndGame={() => endGame.mutate()}
              startGameError={startGameError}
            />
          )}
        </div>

        {/* Sidebar */}
        <div style={{ background: '#1a2a40', borderRadius: '12px', padding: '1rem', border: '1px solid #2d4a6a', alignSelf: 'start' }}>
          {leaderboard && leaderboard.length > 0
            ? <Leaderboard
                entries={leaderboard}
                dealerUserId={game.dealerUserId}
                currentUserId={user?.id}
                canKick={isAdmin}
                onKick={(pid) => {
                  const entry = leaderboard.find(e => e.playerId === pid)
                  if (entry?.userId === game.dealerUserId) {
                    if (!window.confirm('This player is the dealer. Kicking them will delete the table for everyone. Proceed?')) return
                    deleteGame.mutate(undefined, { onSuccess: () => navigate('/') })
                  } else {
                    leaveGame.mutate(pid)
                  }
                }}
              />
            : <p style={{ color: '#4a6a8a' }}>No players yet</p>
          }
        </div>

      </div>
    </div>
  )
}
