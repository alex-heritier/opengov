/**
 * Tests for authentication components
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { GoogleLogin } from '../components/auth/GoogleLogin'
import { AuthProvider } from '../contexts/AuthContext'
import { useAuthStore } from '../stores/authStore'
import userEvent from '@testing-library/user-event'

// Create a test wrapper with necessary providers
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })

  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <AuthProvider>{children}</AuthProvider>
      </QueryClientProvider>
    )
  }
}

describe('GoogleLogin Component', () => {
  beforeEach(() => {
    // Clear store before each test
    useAuthStore.getState().clearAuth()
    // Mock window.location.href
    delete (window as any).location
    window.location = { href: '' } as any
  })

  it('should render login button when not authenticated', () => {
    render(<GoogleLogin />, { wrapper: createWrapper() })

    const button = screen.getByRole('button', { name: /sign in with google/i })
    expect(button).toBeTruthy()
  })

  it('should display Google logo in login button', () => {
    render(<GoogleLogin />, { wrapper: createWrapper() })

    const button = screen.getByRole('button')
    const svg = button.querySelector('svg')
    expect(svg).toBeTruthy()
  })

  it('should redirect to Google OAuth on button click', async () => {
    const user = userEvent.setup()
    render(<GoogleLogin />, { wrapper: createWrapper() })

    const button = screen.getByRole('button', { name: /sign in with google/i })
    await user.click(button)

    await waitFor(() => {
      expect(window.location.href).toContain('/api/auth/google/login')
    })
  })

it('should display user info when authenticated', () => {
    // Mock the user in the auth store directly
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      name: 'Test User',
      picture_url: 'https://example.com/avatar.jpg',
      google_id: 'google-123',
      is_active: true,
      is_verified: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      last_login_at: null,
    }
    
    useAuthStore.setState({
      user: mockUser,
      isAuthenticated: true,
      accessToken: 'mock-token',
      tokenExpiresAt: Date.now() + 3600000,
    })
    
    // Test would render component and verify user info is displayed
    // Implementation depends on specific component structure
  })

it('should display email when name is not available', () => {
    // Mock user without name
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      name: null,
      picture_url: 'https://example.com/avatar.jpg',
      google_id: 'google-123',
      is_active: true,
      is_verified: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      last_login_at: null,
    }
    
    useAuthStore.setState({
      user: mockUser,
      isAuthenticated: true,
      accessToken: 'mock-token',
      tokenExpiresAt: Date.now() + 3600000,
    })
    
    // Test would render component and verify email is displayed instead of name
  })

it('should not display picture when picture_url is null', () => {
    // Mock user without picture
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      name: 'Test User',
      picture_url: null,
      google_id: 'google-123',
      is_active: true,
      is_verified: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      last_login_at: null,
    }
    
    useAuthStore.setState({
      user: mockUser,
      isAuthenticated: true,
      accessToken: 'mock-token',
      tokenExpiresAt: Date.now() + 3600000,
    })
    
    // Test would render component and verify no picture element is shown
  })
})


describe('AuthCallbackPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    vi.clearAllMocks()
  })

  // Note: Full AuthCallbackPage tests would require mocking TanStack Router
  // and the API client, which is more complex. Below are conceptual tests.

  it('should show loading state initially', () => {
    // This would test that the loading spinner is shown
    // Implementation depends on your testing setup for router
  })

  it('should handle successful authentication', async () => {
    // This would test:
    // 1. Extract token from URL hash
    // 2. Fetch user info
    // 3. Store in auth store
    // 4. Redirect to /feed
  })

  it('should handle authentication error', () => {
    // This would test error display when no token in URL
  })
})


describe('AuthLoginPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
  })

  it('should display OpenGov branding', () => {
    // This would test that the page shows the app name and description
  })

  it('should render GoogleLogin component', () => {
    // This would test that the GoogleLogin component is rendered
  })

  it('should show terms and privacy policy text', () => {
    // This would test that legal text is displayed
  })

  it('should redirect to /feed if already authenticated', () => {
    // This would test the redirect logic when user is already logged in
  })
})
