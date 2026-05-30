import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'

const API_BASE = (import.meta.env.VITE_API_URL ?? 'http://localhost:8080')
  .replace(/^http/, 'ws') // http://... → ws://..., https://... → wss://...

type GameEvent =
  | 'game_started'
  | 'game_ended'
  | 'cards_dealt'
  | 'player_joined'
  | 'player_left'
  | 'shoe_shuffled'

/**
 * Opens a WebSocket connection to the game hub and invalidates the relevant
 * TanStack Query caches on each push event.
 * Polling stays active as a slow fallback in case the socket drops.
 */
export function useGameSocket(gameId: string) {
  const qc = useQueryClient()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) return

    const ws = new WebSocket(`${API_BASE}/games/${gameId}/ws?token=${token}`)

    ws.onmessage = (e: MessageEvent) => {
      const { event } = JSON.parse(e.data) as { event: GameEvent }

      switch (event) {
        case 'cards_dealt':
        case 'shoe_shuffled':
          qc.invalidateQueries({ queryKey: ['leaderboard', gameId] })
          qc.invalidateQueries({ queryKey: ['game', gameId] })
          qc.invalidateQueries({ queryKey: ['suits', gameId] })
          qc.invalidateQueries({ queryKey: ['cards', gameId] })
          qc.refetchQueries({ queryKey: ['hand'] })
          break
        case 'player_joined':
        case 'player_left':
          qc.invalidateQueries({ queryKey: ['leaderboard', gameId] })
          qc.invalidateQueries({ queryKey: ['game', gameId] })
          break
        case 'game_started':
        case 'game_ended':
          qc.invalidateQueries({ queryKey: ['game', gameId] })
          qc.invalidateQueries({ queryKey: ['games'] })
          break
      }
    }

    return () => ws.close()
  }, [gameId, qc])
}
