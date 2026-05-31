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

interface UseGameSocketOptions {
  onGameDeleted?: () => void
  onKicked?: () => void
  onNotEnoughPlayers?: () => void
  currentUserId?: string
}

/**
 * Opens a WebSocket connection to the game hub and invalidates the relevant
 * TanStack Query caches on each push event.
 * Polling stays active as a slow fallback in case the socket drops.
 */
export function useGameSocket(gameId: string, options: UseGameSocketOptions = {}) {
  const qc = useQueryClient()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) return

    const ws = new WebSocket(`${API_BASE}/games/${gameId}/ws?token=${token}`)

    ws.onmessage = (e: MessageEvent) => {
      const { event, payload } = JSON.parse(e.data) as {
        event: GameEvent
        payload?: { reason?: string; userId?: string }
      }

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
          qc.invalidateQueries({ queryKey: ['leaderboard', gameId] })
          qc.invalidateQueries({ queryKey: ['game', gameId] })
          break
        case 'player_left':
          qc.invalidateQueries({ queryKey: ['leaderboard', gameId] })
          qc.invalidateQueries({ queryKey: ['game', gameId] })
          // If the event carries a userId and it matches the current user,
          // they were kicked — show a message and redirect to lobby.
          if (payload?.userId && payload.userId === options.currentUserId) {
            options.onKicked?.()
          }
          break
        case 'game_ended':
          if (payload?.reason === 'deleted') {
            options.onGameDeleted?.()
          } else if (payload?.reason === 'not_enough_players') {
            options.onNotEnoughPlayers?.()
          } else {
            qc.invalidateQueries({ queryKey: ['game', gameId] })
            qc.invalidateQueries({ queryKey: ['games'] })
          }
          break
        case 'game_started':
          qc.invalidateQueries({ queryKey: ['game', gameId] })
          qc.invalidateQueries({ queryKey: ['games'] })
          qc.invalidateQueries({ queryKey: ['leaderboard', gameId] })
          qc.invalidateQueries({ queryKey: ['suits', gameId] })
          qc.invalidateQueries({ queryKey: ['cards', gameId] })
          qc.refetchQueries({ queryKey: ['hand'] })
          break
      }
    }

    return () => ws.close()
  }, [gameId, qc]) // options intentionally omitted — callback identity not stable
}
