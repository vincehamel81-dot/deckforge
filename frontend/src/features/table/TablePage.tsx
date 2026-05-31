import { useState, useEffect, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useRequireAuth } from '../../shared/hooks/useRequireAuth'
import { useAuthStore } from '../auth/authStore'
import {
  useGameDetail, useLeaderboard, usePlayerHand, useSuitCounts, useCardCounts,
  useStartGame, useEndGame, useShuffle, useDealToAll, useAddDeck, useLeaveGame, useDeleteGame,
} from './useTable'
import { useGameSocket } from './useGameSocket'
import { CardBadge, SUIT_SYMBOL, SUIT_COLOR } from './CardBadge'
import { Leaderboard } from './Leaderboard'
import { GameResult } from './GameResult'
import { DealerControls } from './DealerControls'
import { LocaleSwitcher } from '../../shared/components/LocaleSwitcher'

export default function TablePage() {
  const authed = useRequireAuth()
  const { t } = useTranslation(['common', 'table', 'errors', 'lobby'])
  const { id: gameId } = useParams<{ id: string }>()
  const user = useAuthStore(s => s.user)
  const logout = useAuthStore(s => s.logout)
  const navigate = useNavigate()

  const [closedReason, setClosedReason] = useState<'deleted' | 'kicked' | 'not_enough_players' | null>(null)
  const [shoeSnapshot, setShoeSnapshot] = useState<ReturnType<typeof useSuitCounts>['data']>(undefined)
  const [isRefreshingSuits, setIsRefreshingSuits] = useState(false)

  useGameSocket(gameId!, {
    onGameDeleted: () => setClosedReason(r => r ?? 'deleted'),
    onKicked: () => setClosedReason(r => r ?? 'kicked'),
    onNotEnoughPlayers: () => setClosedReason(r => r ?? 'not_enough_players'),
    currentUserId: user?.id,
  })

  const { data: detail, isLoading, isError } = useGameDetail(gameId!)
  const { data: leaderboard } = useLeaderboard(gameId!)
  const myEntry = leaderboard?.find(e => e.userId === user?.id)
  const { data: hand } = usePlayerHand(gameId!, myEntry?.playerId)
  const { data: suitCounts, refetch: refetchSuits } = useSuitCounts(gameId!)
  useCardCounts(gameId!) // kept for cache warm-up; data surfaced via shoe queries

  // Populate the snapshot on first load without marking it as stale.
  useEffect(() => {
    if (suitCounts && !shoeSnapshot) {
      setShoeSnapshot(suitCounts)
    }
  }, [suitCounts, shoeSnapshot])

  // Derived: snapshot is stale when live suit counts differ from the last refresh.
  // Comparing data eliminates the need for WS-event-driven shoeStale state, which
  // was unreliable because the options closure in useGameSocket is intentionally
  // excluded from the dependency array.
  const shoeIsStale = useMemo(() => {
    if (!suitCounts || !shoeSnapshot) return false
    return suitCounts.some(s => {
      const snap = shoeSnapshot.find(ss => ss.suit === s.suit)
      return !snap || snap.count !== s.count
    })
  }, [suitCounts, shoeSnapshot])

  const startGame = useStartGame(gameId!)
  const endGame = useEndGame(gameId!)
  const shuffle = useShuffle(gameId!)
  const dealToAll = useDealToAll(gameId!)
  const addDeck = useAddDeck(gameId!)
  const leaveGame = useLeaveGame(gameId!)
  const deleteGame = useDeleteGame(gameId!)

  if (!authed) return null

  const topBar = (
    <div style={{
      position: 'sticky', top: 0, zIndex: 10,
      background: '#0f1a2e', borderBottom: '1px solid #1a2a40',
      padding: '0.75rem 1.5rem',
      display: 'flex', justifyContent: 'space-between', alignItems: 'center',
    }}>
      <span style={{ color: '#e2c97e', fontWeight: 700, fontSize: '1.3rem' }}>{t('appName')}</span>
      <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
        <span style={{ color: '#7a9bb5', fontSize: '0.85rem' }}>👤 {user?.username}</span>
        <LocaleSwitcher />
        <button
          onClick={() => { logout(); navigate('/login') }}
          style={{ padding: '0.4rem 0.8rem', background: 'transparent', border: '1px solid #4a6a8a', borderRadius: '6px', color: '#7a9bb5', cursor: 'pointer' }}
        >{t('logout')}</button>
      </div>
    </div>
  )

  if (closedReason || isError) return (
    <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#e2e8f0', display: 'flex', flexDirection: 'column' }}>
      {topBar}
      <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <div style={{ background: '#1a2a40', borderRadius: '16px', padding: '2.5rem', maxWidth: '400px', width: '100%', textAlign: 'center', border: '1px solid #f8717144' }}>
          <div style={{ fontSize: '2.5rem', marginBottom: '0.5rem' }}>🚪</div>
          <h2 style={{ color: '#f87171', margin: '0 0 0.75rem' }}>
            {closedReason === 'kicked' ? t('table:kicked') : t('table:tableClosed')}
          </h2>
          <p style={{ color: '#7a9bb5', marginBottom: '1.5rem' }}>
            {closedReason === 'kicked'
              ? t('table:kickedMessage')
              : closedReason === 'not_enough_players'
                ? t('table:notEnoughPlayersMessage')
                : t('table:tableClosedMessage')}
          </p>
          <button
            onClick={() => navigate('/', { replace: true })}
            style={{ padding: '0.75rem 2rem', background: '#e2c97e', color: '#0f1a2e', border: 'none', borderRadius: '8px', fontWeight: 700, cursor: 'pointer', width: '100%' }}
          >
            {t('returnToLobby')}
          </button>
        </div>
      </div>
    </div>
  )

  if (isLoading || !detail) return (
    <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#7a9bb5', display: 'flex', flexDirection: 'column' }}>
      {topBar}
      <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        {t('table:loadingTable')}
      </div>
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
      <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#7a9bb5', display: 'flex', flexDirection: 'column' }}>
        {topBar}
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          {t('table:loadingResults')}
        </div>
      </div>
    )
    return (
      <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#e2e8f0', display: 'flex', flexDirection: 'column' }}>
        {topBar}
        <GameResult
          leaderboard={leaderboard}
          dealerUserId={game.dealerUserId}
          currentUserId={user?.id}
          gameId={gameId!}
          onLobby={() => navigate('/')}
        />
      </div>
    )
  }

  const startGameError = startGame.isError
    ? t('errors:COULD_NOT_START')
    : undefined

  const suitLabel = (suit: string) => {
    const map: Record<string, string> = {
      HEARTS: t('table:shoe.hearts'),
      SPADES: t('table:shoe.spades'),
      CLUBS: t('table:shoe.clubs'),
      DIAMONDS: t('table:shoe.diamonds'),
    }
    return map[suit] ?? suit
  }

  const handleRefreshSuits = async () => {
    setIsRefreshingSuits(true)
    const result = await refetchSuits()
    if (result.data) setShoeSnapshot(result.data)
    setIsRefreshingSuits(false)
  }

  return (
    <div style={{ minHeight: '100vh', background: '#0f1a2e', color: '#e2e8f0' }}>

      {/* Sticky header */}
      <div style={{
        position: 'sticky', top: 0, zIndex: 10,
        background: '#0f1a2e', borderBottom: '1px solid #1a2a40',
        padding: '0.75rem 1.5rem',
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
      }}>
        <div>
          <span style={{ color: '#e2c97e', fontWeight: 700, fontSize: '1.3rem' }}>{t('appName')}</span>
          <span style={{ color: '#4a6a8a', marginLeft: '1rem', fontSize: '0.85rem' }}>
            {t('table:tableCreatedBy', { dealer: displayDealerName })}
          </span>
        </div>
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <span style={{ color: '#7a9bb5', fontSize: '0.85rem' }}>👤 {user?.username}</span>
          <span style={{
            padding: '0.2rem 0.7rem', borderRadius: '4px', fontSize: '0.8rem', fontWeight: 600,
            background: game.status === 'WAITING' ? '#1a4a2a' : '#1a2a4a',
            color: game.status === 'WAITING' ? '#4ade80' : '#60a5fa',
          }}>
            {t(`table:status.${game.status}`)}
          </span>
          <LocaleSwitcher />
          <button
            onClick={async () => {
              if (myEntry && game.status !== 'FINISHED') {
                if (isDealer) {
                  if (!window.confirm(t('lobby:confirmLeaveDealer'))) return
                  await deleteGame.mutateAsync()
                  navigate('/')
                  return
                }
                if (!window.confirm(t('lobby:confirmLeavePlayer'))) return
                await leaveGame.mutateAsync(myEntry.playerId)
              }
              navigate('/')
            }}
            style={{ padding: '0.4rem 0.8rem', background: 'transparent', border: '1px solid #4a6a8a', borderRadius: '6px', color: '#7a9bb5', cursor: 'pointer' }}
          >{t('toLobby')}</button>
          <button
            onClick={() => { logout(); navigate('/login') }}
            style={{ padding: '0.4rem 0.8rem', background: 'transparent', border: '1px solid #4a6a8a', borderRadius: '6px', color: '#7a9bb5', cursor: 'pointer' }}
          >{t('logout')}</button>
        </div>
      </div>

      <div style={{ padding: '1.5rem' }}>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 300px', gap: '1.5rem' }}>

        {/* Main column */}
        <div>
          {/* Shoe status */}
          <div style={{ background: '#1a2a40', borderRadius: '12px', padding: '1rem', marginBottom: '1rem', border: '1px solid #2d4a6a' }}>
            <h3 style={{ color: '#e2c97e', margin: '0 0 0.75rem' }}>{t('table:shoe.title')}</h3>
            <div style={{ display: 'flex', gap: '2rem', flexWrap: 'wrap', marginBottom: '0.75rem' }}>
              <div><span style={{ color: '#4a6a8a' }}>{t('table:shoe.total')} </span><strong>{totalCards}</strong></div>
              <div><span style={{ color: '#4a6a8a' }}>{t('table:shoe.remaining')} </span><strong style={{ color: remainingCards < 10 ? '#f87171' : '#4ade80' }}>{remainingCards}</strong></div>
              <div><span style={{ color: '#4a6a8a' }}>{t('table:shoe.drawsLeft')} </span><strong style={{ color: '#e2c97e' }}>{drawsRemaining}</strong></div>
              <div><span style={{ color: '#4a6a8a' }}>{t('table:shoe.decks')} </span><strong>{game.deckCount}</strong></div>
            </div>
            {suitCounts && (
              <div style={{ display: 'flex', gap: '1.5rem', flexWrap: 'wrap', marginBottom: '0.75rem' }}>
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
            {/* Undealt per suit — always visible; refresh button appears when shoe state changed */}
            <div style={{ borderTop: '1px solid #2d4a6a', paddingTop: '0.75rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.4rem' }}>
                <span style={{ color: '#4a6a8a', fontSize: '0.75rem' }}>{t('table:shoe.undealtBySuit')}</span>
                {shoeIsStale && (
                  <button
                    onClick={handleRefreshSuits}
                    disabled={isRefreshingSuits}
                    title={t('table:shoe.refreshShoeTitle')}
                    data-tooltip={t('table:shoe.refreshShoeTitle')}
                    style={{ padding: '0.1rem 0.4rem', background: 'transparent', border: '1px solid #4a6a8a', borderRadius: '3px', color: '#7a9bb5', cursor: isRefreshingSuits ? 'default' : 'pointer', fontSize: '0.7rem' }}
                  >
                    {isRefreshingSuits ? t('table:shoe.checking') : t('table:shoe.refreshShoe')}
                  </button>
                )}
              </div>
              {shoeSnapshot ? (
                <div style={{ display: 'flex', gap: '1.5rem', flexWrap: 'wrap' }}>
                  {shoeSnapshot.map(s => (
                    <span key={s.suit} style={{ fontSize: '0.85rem' }}>
                      <span style={{ color: SUIT_COLOR[s.suit] ?? '#e2e8f0' }}>{suitLabel(s.suit)}</span>
                      <span style={{ color: '#e2e8f0' }}>: <strong>{s.count}</strong></span>
                    </span>
                  ))}
                </div>
              ) : (
                <span style={{ color: '#4a6a8a', fontSize: '0.8rem' }}>…</span>
              )}
            </div>
          </div>

          {/* Warn when shoe can't serve a full round */}
          {game.status === 'IN_PROGRESS' && leaderboard && leaderboard.length > 0 && remainingCards < leaderboard.length && (
            <div style={{ background: '#2a1a1a', border: '1px solid #f8717166', borderRadius: '8px', padding: '0.6rem 0.9rem', marginBottom: '1rem', fontSize: '0.85rem', color: '#f87171' }}>
              {remainingCards === 0
                ? t('table:shoe.exhausted')
                : t('table:shoe.lowCards', { count: remainingCards })}
            </div>
          )}

          {/* Player hand */}
          <div style={{ background: '#1a2a40', borderRadius: '12px', padding: '1rem', marginBottom: '1rem', border: '1px solid #2d4a6a' }}>
            <h3 style={{ color: '#e2c97e', marginBottom: '0.75rem' }}>
              {t('table:hand.title')} {myEntry && <span style={{ color: '#7a9bb5', fontWeight: 400 }}>{t('table:hand.pts', { value: myEntry.handValue })}</span>}
            </h3>
            {!hand || hand.length === 0
              ? <p style={{ color: '#4a6a8a' }}>{t('table:hand.noCards')}</p>
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
                    if (!window.confirm(t('lobby:confirmKickDealer'))) return
                    deleteGame.mutate(undefined, { onSuccess: () => navigate('/') })
                  } else {
                    leaveGame.mutate(pid)
                  }
                }}
              />
            : <p style={{ color: '#4a6a8a' }}>{t('table:leaderboard.noPlayers')}</p>
          }
        </div>

      </div>
    </div>
    </div>
  )
}
