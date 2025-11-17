import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { jwtDecode } from 'jwt-decode'

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

interface JWTPayload {
  exp: number
  [key: string]: any
}

interface AuthState {
  user: User | null
  accessToken: string | null
  tokenExpiresAt: number | null
  isAuthenticated: boolean

  // Actions
  setAuth: (accessToken: string, user: User) => void
  clearAuth: () => void
  isTokenExpiringSoon: () => boolean
}

// Calculate token expiration from JWT
const getTokenExpiration = (token: string): number => {
  try {
    const decoded = jwtDecode<JWTPayload>(token)
    return decoded.exp * 1000 // Convert to milliseconds
  } catch {
    return 0
  }
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      tokenExpiresAt: null,
      isAuthenticated: false,

      setAuth: (accessToken: string, user: User) => {
        const expiresAt = getTokenExpiration(accessToken)
        set({
          user,
          accessToken,
          tokenExpiresAt: expiresAt,
          isAuthenticated: true,
        })
      },

      clearAuth: () => {
        set({
          user: null,
          accessToken: null,
          tokenExpiresAt: null,
          isAuthenticated: false,
        })
      },

      isTokenExpiringSoon: () => {
        const { tokenExpiresAt } = get()
        if (!tokenExpiresAt) return false

        const now = Date.now()
        const timeLeft = tokenExpiresAt - now

        // Consider token expiring soon if less than 10 minutes left
        const tenMinutesInMs = 10 * 60 * 1000
        return timeLeft < tenMinutesInMs && timeLeft > 0
      },
    }),
    {
      name: 'opengov-auth',
      storage: createJSONStorage(() => localStorage),
    }
  )
)
