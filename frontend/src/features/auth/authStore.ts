import { create } from 'zustand'
import apiClient from '../../lib/apiClient'

interface AuthUser {
  id: string
  username: string
  role: string
}

interface AuthState {
  user: AuthUser | null
  token: string | null
  login: (username: string) => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: (() => {
    try {
      const raw = localStorage.getItem('user')
      return raw ? JSON.parse(raw) : null
    } catch {
      return null
    }
  })(),
  token: localStorage.getItem('token'),

  login: async (username: string) => {
    // Try register first; fall back to login on 409 (username taken)
    let response
    try {
      response = await apiClient.post('/auth/register', { username })
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status
      if (status === 409) {
        response = await apiClient.post('/auth/login', { username })
      } else {
        throw err
      }
    }
    const { token, user } = response.data
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    set({ token, user })
  },

  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    set({ token: null, user: null })
  },
}))
