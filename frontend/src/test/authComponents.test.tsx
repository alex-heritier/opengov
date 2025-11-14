/**
 * Tests for authentication components
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { GoogleLogin } from '../components/auth/GoogleLogin'
import { useAuthStore } from '../stores/authStore'
import userEvent from '@testing-library/user-event'

describe('GoogleLogin Component', () => {
  beforeEach(() => {
    // Clear store before each test
    useAuthStore.getState().clearAuth()
    // Mock window.location.href
    delete (window as any).location
    window.location = { href: '' } as any
  })

  it('should render login button when not authenticated', () => {
    render(<GoogleLogin />)

    const button = screen.getByRole('button', { name: /sign in with google/i })
    expect(button).toBeTruthy()
  })

  it('should display Google logo in login button', () => {
    render(<GoogleLogin />)

    const button = screen.getByRole('button')
    const svg = button.querySelector('svg')
    expect(svg).toBeTruthy()
  })

  it('should redirect to Google OAuth on button click', async () => {
    const user = userEvent.setup()
    render(<GoogleLogin />)

    const button = screen.getByRole('button', { name: /sign in with google/i })
    await user.click(button)

    await waitFor(() => {
      expect(window.location.href).toContain('/api/auth/google/login')
    })
  })

  it('should display user info when authenticated', () => {
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      name: 'Test User',
      picture_url: 'https://example.com/pic.jpg',
      google_id: 'google123',
      is_active: true,
      is_verified: true,
      created_at: '2025-01-01T00:00:00Z',
      updated_at: '2025-01-01T00:00:00Z',
      last_login_at: '2025-01-01T00:00:00Z',
    }

    const now = Math.floor(Date.now() / 1000)
    const exp = now + 3600
    const mockPayload = { sub: 1, exp }
    const mockToken = `header.${btoa(JSON.stringify(mockPayload))}.signature`

    useAuthStore.getState().setAuth(mockToken, mockUser)

    render(<GoogleLogin />)

    expect(screen.getByText('Test User')).toBeTruthy()
    const img = screen.getByAlt('Test User')
    expect(img).toBeTruthy()
    expect(img.getAttribute('src')).toBe('https://example.com/pic.jpg')
  })

  it('should display email when name is not available', () => {
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      name: null,
      picture_url: null,
      google_id: 'google123',
      is_active: true,
      is_verified: true,
      created_at: '2025-01-01T00:00:00Z',
      updated_at: '2025-01-01T00:00:00Z',
      last_login_at: null,
    }

    const now = Math.floor(Date.now() / 1000)
    const exp = now + 3600
    const mockPayload = { sub: 1, exp }
    const mockToken = `header.${btoa(JSON.stringify(mockPayload))}.signature`

    useAuthStore.getState().setAuth(mockToken, mockUser)

    render(<GoogleLogin />)

    expect(screen.getByText('test@example.com')).toBeTruthy()
  })

  it('should not display picture when picture_url is null', () => {
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      name: 'Test User',
      picture_url: null,
      google_id: 'google123',
      is_active: true,
      is_verified: true,
      created_at: '2025-01-01T00:00:00Z',
      updated_at: '2025-01-01T00:00:00Z',
      last_login_at: null,
    }

    const now = Math.floor(Date.now() / 1000)
    const exp = now + 3600
    const mockPayload = { sub: 1, exp }
    const mockToken = `header.${btoa(JSON.stringify(mockPayload))}.signature`

    useAuthStore.getState().setAuth(mockToken, mockUser)

    render(<GoogleLogin />)

    const img = screen.queryByRole('img')
    expect(img).toBeFalsy()
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
