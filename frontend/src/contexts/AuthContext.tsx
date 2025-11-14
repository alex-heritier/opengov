/**
 * Auth context provider using React Query for server state
 * Separates server state (user data) from client state (token, UI)
 */
import { createContext, useContext, ReactNode, useCallback, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '../stores/authStore'
import client from '../api/client'

interface User {
  id: number
  email: string
  name: string | null
  picture_url: string | null
  google_id: string | null
  is_active: boolean
  is_verified: boolean
  created_at: string
  updated_at: string
  last_login_at: string | null
}

interface AuthContextValue {
  user: User | null
  isLoading: boolean
  isError: boolean
  error: Error | null
  isAuthenticated: boolean
  login: () => void
  logout: () => void
  renewToken: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

// API functions
async function fetchCurrentUser(): Promise<User> {
  const response = await client.get('/api/auth/me')
  return response.data
}

async function renewAccessToken(): Promise<{ access_token: string; expires_in: number }> {
  const response = await client.post('/api/auth/renew')
  return response.data
}

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const queryClient = useQueryClient()
  const { accessToken, clearAuth, setAuth, isTokenExpiringSoon } = useAuthStore()

  // Query for current user (only runs if accessToken exists)
  const {
    data: user,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ['user', accessToken],
    queryFn: fetchCurrentUser,
    enabled: !!accessToken,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (failureCount, error: any) => {
      // Don't retry on 401 errors
      if (error?.response?.status === 401) {
        return false
      }
      return failureCount < 2
    },
  })

  // Mutation for token renewal
  const renewMutation = useMutation({
    mutationFn: renewAccessToken,
    onSuccess: async (data) => {
      // Fetch updated user info with new token
      const response = await client.get('/api/auth/me', {
        headers: {
          Authorization: `Bearer ${data.access_token}`,
        },
      })
      const updatedUser = response.data

      // Update auth store with new token and user
      setAuth(data.access_token, updatedUser)

      // Invalidate user query to refetch with new token
      queryClient.invalidateQueries({ queryKey: ['user'] })
    },
    onError: () => {
      // Token renewal failed - clear auth
      clearAuth()
      queryClient.clear()
    },
  })

  // Auto-renew token when expiring soon
  useEffect(() => {
    if (!accessToken) return

    const interval = setInterval(() => {
      if (isTokenExpiringSoon() && !renewMutation.isPending) {
        renewMutation.mutate()
      }
    }, 60000) // Check every minute

    return () => clearInterval(interval)
  }, [accessToken, isTokenExpiringSoon, renewMutation])

  const login = useCallback(() => {
    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8000'
    window.location.href = `${apiUrl}/api/auth/google/login`
  }, [])

  const logout = useCallback(() => {
    clearAuth()
    queryClient.clear()
    window.location.href = '/auth/login'
  }, [clearAuth, queryClient])

  const renewToken = useCallback(async () => {
    await renewMutation.mutateAsync()
  }, [renewMutation])

  const value: AuthContextValue = {
    user: user || null,
    isLoading,
    isError,
    error: error as Error | null,
    isAuthenticated: !!accessToken && !!user,
    login,
    logout,
    renewToken,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
