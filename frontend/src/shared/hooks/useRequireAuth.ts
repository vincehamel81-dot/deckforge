import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../features/auth/authStore'

export function useRequireAuth() {
  const token = useAuthStore((s) => s.token)
  const navigate = useNavigate()
  useEffect(() => {
    if (!token) navigate('/login', { replace: true })
  }, [token, navigate])
  return !!token
}
