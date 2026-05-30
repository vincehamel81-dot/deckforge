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
    // Try login first; register automatically if the username doesn't exist yet (404).
    let response
    try {
      response = await apiClient.post('/auth/login', { username })
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status
      if (status === 404) {
        response = await apiClient.post('/auth/register', { username })
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
