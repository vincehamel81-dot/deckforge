import { useState } from 'react'
import { useTranslation } from 'react-i18next'

interface DealerControlsProps {
  gameStatus: 'WAITING' | 'IN_PROGRESS'
  playerCount: number
  minPlayers: number
  onAddDeck: () => void
  onShuffle: () => Promise<void>
  onStartGame: (initialDeal: number) => void
  onDealToAll: (count: number) => void
  onEndGame: () => void
  startGameError?: string
}

export function DealerControls({
  gameStatus, playerCount, minPlayers,
  onAddDeck, onShuffle, onStartGame, onDealToAll, onEndGame, startGameError,
}: DealerControlsProps) {
  const { t } = useTranslation('table')
  const [dealCount, setDealCount] = useState(1)
  const [startDeal, setStartDeal] = useState(2)
  const [shuffleMsg, setShuffleMsg] = useState('')

  async function handleShuffle() {
    await onShuffle()
    setShuffleMsg(t('dealer.shuffled'))
    setTimeout(() => setShuffleMsg(''), 2500)
  }

  return (
    <div style={{ background: '#1a2a40', borderRadius: '12px', padding: '1rem', border: '1px solid #e2c97e44' }}>
      <h3 style={{ color: '#e2c97e', marginBottom: '1rem' }}>{t('dealer.title')}</h3>
      <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', alignItems: 'center' }}>

        {gameStatus === 'WAITING' && (
          <>
            <button onClick={onAddDeck} style={{ padding: '0.5rem 1rem', background: '#1a3a4a', color: '#60a5fa', border: '1px solid #60a5fa', borderRadius: '8px', cursor: 'pointer' }}>
              {t('dealer.addDeck')}
            </button>
            <button onClick={handleShuffle} style={{ padding: '0.5rem 1rem', background: '#1a2a4a', color: '#a78bfa', border: '1px solid #a78bfa', borderRadius: '8px', cursor: 'pointer' }}>
              {t('dealer.shuffle')}
            </button>
            {shuffleMsg && <span style={{ color: '#a78bfa', fontSize: '0.85rem' }}>{shuffleMsg}</span>}
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span style={{ color: '#7a9bb5', fontSize: '0.85rem' }}>{t('dealer.initialDeal')}</span>
              <input
                type="number" min={0} max={10} value={startDeal}
                onChange={e => setStartDeal(+e.target.value)}
                style={{ width: '50px', padding: '0.4rem', background: '#0f1a2e', border: '1px solid #2d4a6a', borderRadius: '6px', color: '#e2e8f0', textAlign: 'center' }}
              />
            </div>
            {playerCount < minPlayers ? (
              <span style={{ color: '#7a9bb5', fontSize: '0.85rem' }}>
                {t('dealer.waitingForPlayers', { current: playerCount, min: minPlayers })}
              </span>
            ) : (
              <button onClick={() => onStartGame(startDeal)} style={{ padding: '0.5rem 1.2rem', background: '#1a4a2a', color: '#4ade80', border: '1px solid #4ade80', borderRadius: '8px', cursor: 'pointer', fontWeight: 700 }}>
                {t('dealer.startGame')}
              </button>
            )}
            {startGameError && (
              <span style={{ color: '#f87171', fontSize: '0.85rem' }}>{startGameError}</span>
            )}
          </>
        )}

        {gameStatus === 'IN_PROGRESS' && (
          <>
            <button onClick={handleShuffle} style={{ padding: '0.5rem 1rem', background: '#1a2a4a', color: '#a78bfa', border: '1px solid #a78bfa', borderRadius: '8px', cursor: 'pointer' }}>
              {t('dealer.shuffle')}
            </button>
            {shuffleMsg && <span style={{ color: '#a78bfa', fontSize: '0.85rem' }}>{shuffleMsg}</span>}
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span style={{ color: '#7a9bb5', fontSize: '0.85rem' }}>{t('dealer.dealCards')}</span>
              <input
                type="number" min={1} max={10} value={dealCount}
                onChange={e => setDealCount(Math.max(1, +e.target.value))}
                style={{ width: '50px', padding: '0.4rem', background: '#0f1a2e', border: '1px solid #2d4a6a', borderRadius: '6px', color: '#e2e8f0', textAlign: 'center' }}
              />
            </div>
            <button onClick={() => onDealToAll(dealCount)} style={{ padding: '0.5rem 1rem', background: '#e2c97e', color: '#0f1a2e', border: 'none', borderRadius: '8px', cursor: 'pointer', fontWeight: 700 }}>
              {t('dealer.dealToAll')}
            </button>
            <button onClick={onEndGame} style={{ padding: '0.5rem 1rem', background: '#2a1a1a', color: '#f87171', border: '1px solid #f87171', borderRadius: '8px', cursor: 'pointer' }}>
              {t('dealer.endGame')}
            </button>
          </>
        )}

      </div>
    </div>
  )
}
