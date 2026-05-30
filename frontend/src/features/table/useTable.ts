import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import apiClient from '../../lib/apiClient'

export interface ShoeCard {
  id: string
  suit: 'HEARTS' | 'SPADES' | 'CLUBS' | 'DIAMONDS'
  face: string
  numericValue: number
}

export interface LeaderboardEntry {
  playerId: string
  userId: string
  username: string
  seatOrder: number
  handValue: number
  cardCount: number
}

export interface GameDetail {
  game: {
    id: string
    dealerUserId: string
    status: 'WAITING' | 'IN_PROGRESS' | 'FINISHED'
    deckCount: number
    minPlayers: number
    maxPlayers: number
  }
  totalCards: number
  remainingCards: number
  dealerUsername: string
}

export function useGameDetail(gameId: string) {
  return useQuery<GameDetail>({
    queryKey: ['game', gameId],
    queryFn: () => apiClient.get(`/games/${gameId}`).then(r => r.data),
    refetchInterval: 15000, // WS push handles real-time; polling is the fallback
  })
}

export function useLeaderboard(gameId: string) {
  return useQuery<LeaderboardEntry[]>({
    queryKey: ['leaderboard', gameId],
    queryFn: () => apiClient.get(`/games/${gameId}/players`).then(r => r.data.leaderboard),
    refetchInterval: 15000,
  })
}

export function usePlayerHand(gameId: string, playerId: string | undefined) {
  return useQuery<ShoeCard[]>({
    queryKey: ['hand', playerId],
    queryFn: () => apiClient.get(`/games/${gameId}/players/${playerId}/hand`).then(r => r.data.cards),
    enabled: !!playerId,
    refetchInterval: 15000,
  })
}

export interface SuitCount {
  suit: string
  count: number
}

export interface CardCount {
  suit: string
  face: string
  count: number
}

export function useSuitCounts(gameId: string) {
  return useQuery<SuitCount[]>({
    queryKey: ['suits', gameId],
    queryFn: () => apiClient.get(`/games/${gameId}/shoe/suits`).then(r => r.data.suits),
    refetchInterval: 15000,
  })
}

export function useCardCounts(gameId: string) {
  return useQuery<CardCount[]>({
    queryKey: ['cards', gameId],
    queryFn: () => apiClient.get(`/games/${gameId}/shoe/cards`).then(r => r.data.cards),
    refetchInterval: 15000,
  })
}

export function useStartGame(gameId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (initialDealCount: number) =>
      apiClient.post(`/games/${gameId}/start`, { initialDealCount }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['game', gameId] })
      qc.invalidateQueries({ queryKey: ['leaderboard', gameId] })
      qc.invalidateQueries({ queryKey: ['suits', gameId] })
      qc.invalidateQueries({ queryKey: ['cards', gameId] })
      qc.refetchQueries({ queryKey: ['hand'] })
    },
  })
}

export function useEndGame(gameId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => apiClient.post(`/games/${gameId}/end`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['game', gameId] })
      qc.invalidateQueries({ queryKey: ['games'] })
    },
  })
}

export function useShuffle(gameId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => apiClient.post(`/games/${gameId}/shoe/shuffle`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['game', gameId] })
      qc.invalidateQueries({ queryKey: ['suits', gameId] })
      qc.invalidateQueries({ queryKey: ['cards', gameId] })
    },
  })
}

export function useDealToAll(gameId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ count }: { count: number }) =>
      apiClient.post(`/games/${gameId}/deal-round`, { count }),
    onSuccess: (res) => {
      // If the shoe was empty nothing changed — skip all downstream refetches.
      if (res.data.totalDealt === 0) return
      qc.invalidateQueries({ queryKey: ['leaderboard', gameId] })
      qc.invalidateQueries({ queryKey: ['game', gameId] })
      qc.invalidateQueries({ queryKey: ['suits', gameId] })
      qc.invalidateQueries({ queryKey: ['cards', gameId] })
      qc.refetchQueries({ queryKey: ['hand'] })
    },
  })
}

export function useAddDeck(gameId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const { data } = await apiClient.post('/decks')
      return apiClient.post(`/games/${gameId}/shoe/decks`, { deckId: data.id })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['game', gameId] }),
  })
}

export function useLeaveGame(gameId: string) {
  return useMutation({
    mutationFn: (playerId: string) =>
      apiClient.delete(`/games/${gameId}/players/${playerId}`),
  })
}

export function useDeleteGame(gameId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => apiClient.delete(`/games/${gameId}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['games'] }),
  })
}
