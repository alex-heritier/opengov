import axios from 'axios'
import { useAuthStore } from '../store/authStore'

const API_URL = (import.meta.env.VITE_API_URL as string) || 'http://localhost:8000'

const client = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Flag to prevent multiple simultaneous token renewals
let isRenewing = false
let renewalPromise: Promise<string> | null = null

// Helper function to redirect to login
function redirectToLogin(): void {
  if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/auth')) {
    window.location.href = '/auth/login'
  }
}

// Function to renew token
async function renewToken(currentToken: string): Promise<string> {
  const response = await axios.post(
    `${API_URL}/api/auth/renew`,
    {},
    {
      headers: {
        Authorization: `Bearer ${currentToken}`,
      },
    }
  )
  return response.data.access_token
}

// Function to fetch current user
async function fetchCurrentUser(token: string) {
  const response = await axios.get(`${API_URL}/api/auth/me`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
  return response.data
}

// Request interceptor - add auth token and handle auto-renewal
client.interceptors.request.use(
  async (config) => {
    const { accessToken, isTokenExpiringSoon, setAuth, clearAuth } = useAuthStore.getState()

    // Skip token renewal for auth endpoints
    const isAuthEndpoint = config.url?.includes('/api/auth')

    if (accessToken && isTokenExpiringSoon() && !isAuthEndpoint) {
      try {
        // If already renewing, wait for that promise
        if (isRenewing && renewalPromise) {
          const newToken = await renewalPromise
          config.headers.Authorization = `Bearer ${newToken}`
          return config
        }

        // Start renewal
        isRenewing = true
        renewalPromise = renewToken(accessToken)
        const newToken = await renewalPromise

        // Fetch updated user info
        const user = await fetchCurrentUser(newToken)

        // Update store
        setAuth(newToken, user)

        // Update request config
        config.headers.Authorization = `Bearer ${newToken}`
      } catch (error) {
        console.error('Token renewal failed:', error)
        clearAuth()
        redirectToLogin()
      } finally {
        isRenewing = false
        renewalPromise = null
      }
    } else if (accessToken) {
      // Just add the token if not expiring soon
      config.headers.Authorization = `Bearer ${accessToken}`
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor - handle auth errors
client.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    // Handle 401 Unauthorized errors
    if (error.response?.status === 401) {
      const { clearAuth } = useAuthStore.getState()
      clearAuth()
      redirectToLogin()
    }

    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

export default client
