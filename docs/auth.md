# Authentication System

## Overview

OpenGov uses **Google OAuth 2.0** for authentication with **JWT tokens** stored in **localStorage**. This provides secure, user-friendly authentication without managing passwords.

## Design Principles

- **Trust Google**: Leverage Google's authentication infrastructure
- **Simplicity**: No refresh tokens, no database token storage - stateless JWT only
- **Sliding Window**: Active users can renew tokens without re-authenticating

## Architecture

### Flow

1. User clicks "Sign in with Google"
2. Redirect to Google OAuth consent screen
3. Google redirects back with authorization code
4. Backend exchanges code for user info, creates/updates user
5. Backend issues JWT token (1 hour expiration)
6. Frontend stores token in localStorage
7. API calls include token in `Authorization: Bearer <token>` header
8. Frontend auto-renews token when 50% expired (~30 min)

### Token Strategy

- **Type**: Single JWT access token (no refresh tokens)
- **Expiration**: 1 hour
- **Storage**: localStorage via Zustand persist
- **Renewal**: Exchange non-expired token for fresh token via `POST /api/auth/renew`
- **Auto-Renewal**: Frontend renews when <10 minutes remaining

## Implementation

### Backend

**Environment Variables:**
```bash
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8000/api/auth/google/callback
JWT_SECRET_KEY=your-random-secret-key-min-32-chars
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=60
FRONTEND_URL=http://localhost:5173
```

**Dependencies:**
```txt
authlib==1.3.0
python-jose[cryptography]==3.3.0
python-multipart==0.0.6
```

**Database (Users only - no tokens table):**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    google_id VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    picture_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    last_login_at TIMESTAMP
);
```

**API Endpoints:**
- `GET /api/auth/google/login` - Initiate OAuth
- `GET /api/auth/google/callback` - Handle OAuth callback
- `POST /api/auth/renew` - Renew token (requires valid token)
- `POST /api/auth/logout` - Client-side only
- `GET /api/auth/me` - Get current user

### Frontend

**Auth Store (Zustand):**
```typescript
interface AuthState {
  user: User | null
  accessToken: string | null
  tokenExpiresAt: number | null
  isAuthenticated: boolean

  setAuth: (accessToken: string, user: User) => void
  clearAuth: () => void
  isTokenExpiringSoon: () => boolean
}

// Persists to localStorage automatically
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({ /* ... */ }),
    { name: 'opengov-auth', storage: createJSONStorage(() => localStorage) }
  )
)
```

**API Client Interceptor:**
```typescript
// Auto-renew token if expiring soon (<10 min)
apiClient.interceptors.request.use(async (config) => {
  const { accessToken, isTokenExpiringSoon, setAuth, clearAuth } = useAuthStore.getState()

  if (accessToken && isTokenExpiringSoon()) {
    try {
      const { data } = await axios.post('/api/auth/renew', {}, {
        headers: { Authorization: `Bearer ${accessToken}` }
      })
      const user = await fetchCurrentUser(data.access_token)
      setAuth(data.access_token, user)
      config.headers.Authorization = `Bearer ${data.access_token}`
    } catch {
      clearAuth()
      window.location.href = '/login'
    }
  } else if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }

  return config
})
```

## Security

**Token Security:**
- 1-hour expiration limits exposure
- Stateless JWT - no database lookups
- HTTPS only in production
- Strong secret key (32+ characters)

**XSS Mitigation:**
- Content Security Policy headers
- Input sanitization
- Short token lifetime

**Headers (Backend):**
```python
response.headers["Content-Security-Policy"] = "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';"
response.headers["X-Frame-Options"] = "DENY"
response.headers["X-Content-Type-Options"] = "nosniff"
```

## Google Cloud Setup

1. Go to https://console.cloud.google.com/
2. Create OAuth client ID (Web application)
3. Add authorized redirect URIs:
   - Dev: `http://localhost:8000/api/auth/google/callback`
   - Prod: `https://yourdomain.com/api/auth/google/callback`
4. Copy Client ID and Secret to `.env`

## Trade-offs

**Advantages:**
- Simple - no token database, no refresh logic
- Fast - no database lookups
- Stateless - easy to scale
- Good UX - users stay logged in, auto-renewal

**Disadvantages:**
- Can't revoke tokens before expiration
- localStorage vulnerable to XSS (mitigated by CSP + short expiration)
- No multi-device logout

## Future Enhancements

- Token blacklist for immediate revocation
- httpOnly cookies for enhanced security
- Refresh tokens for longer sessions
