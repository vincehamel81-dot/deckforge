import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import apiClient from '../../lib/apiClient'
import { useAuthStore } from '../auth/authStore'

export interface Game {
  id: string
  dealerUserId: string
  status: 'WAITING' | 'IN_PROGRESS' | 'FINISHED'
  deckCount: number
  minPlayers: number
  maxPlayers: number
  createdAt: string
  playerCount: number
  dealerUsername: string
}

export function useGames() {
  const token = useAuthStore((s) => s.token)
  return useQuery<Game[]>({
    queryKey: ['games'],
    queryFn: async () => {
      const res = await apiClient.get('/games')
      return res.data.games ?? []
    },
    refetchInterval: 5000,
    enabled: !!token,
  })
}

export function useCreateGame() {
  const qc = useQueryClient()
  const navigate = useNavigate()
  return useMutation({
    mutationFn: async (body: { deckCount: number; minPlayers: number; maxPlayers: number }) => {
      const res = await apiClient.post('/games', body)
      const gameId = res.data.game.id
      // Populate the shoe — backend initialises DeckCount at 0; each call adds 52 cards.
      for (let i = 0; i < body.deckCount; i++) {
        const { data: deck } = await apiClient.post('/decks')
        await apiClient.post(`/games/${gameId}/shoe/decks`, { deckId: deck.id })
      }
      return res
    },
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ['games'] })
      navigate(`/table/${res.data.game.id}`)
    },
  })
}

export function useJoinGame() {
  const navigate = useNavigate()
  return useMutation({
    mutationFn: (gameId: string) => apiClient.post(`/games/${gameId}/players`),
    onSuccess: (_, gameId) => navigate(`/table/${gameId}`),
  })
}
