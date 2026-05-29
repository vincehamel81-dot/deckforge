import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import apiClient from '../../lib/apiClient'

export interface Game {
  id: string
  dealerUserId: string
  status: 'WAITING' | 'IN_PROGRESS' | 'FINISHED'
  deckCount: number
  minPlayers: number
  maxPlayers: number
  createdAt: string
}

export function useGames() {
  return useQuery<Game[]>({
    queryKey: ['games'],
    queryFn: async () => {
      const res = await apiClient.get('/games')
      return res.data.games ?? []
    },
    refetchInterval: 5000,
  })
}

export function useCreateGame() {
  const qc = useQueryClient()
  const navigate = useNavigate()
  return useMutation({
    mutationFn: (body: { deckCount: number; minPlayers: number; maxPlayers: number }) =>
      apiClient.post('/games', body),
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
