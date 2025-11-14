# Authentication System

## Overview

OpenGov uses **Google OAuth 2.0** for authentication combined with **JWT tokens** for session management. This provides secure, user-friendly authentication without managing passwords.

## Design Principles

- **Trust Google**: Leverage Google's secure authentication infrastructure
- **Simplicity**: No refresh tokens, no database token storage - just stateless JWT
- **User Experience**: One-click login with Google accounts
- **Sliding Window**: Active users can renew tokens without re-authenticating

## Architecture

### Authentication Flow

1. User clicks "Sign in with Google" button
2. Frontend redirects to `/api/auth/google/login`
3. Backend redirects to Google OAuth consent screen
4. User approves access
5. Google redirects to `/api/auth/google/callback?code=...`
6. Backend exchanges code for user info
7. Backend creates/updates user in database
8. Backend generates JWT token (1-hour expiration)
9. Backend redirects to frontend with token in URL
10. Frontend stores token in **localStorage** and fetches user profile
11. Subsequent API calls include token in `Authorization: Bearer <token>` header

### Token Strategy (Sliding Window)

- **Access Token**: JWT with 1-hour expiration
- **Token Renewal**: Users can exchange non-expired token for fresh token via `POST /api/auth/renew`
- **Active Sessions**: Frontend automatically renews token every 30 minutes
- **Inactive Sessions**: Token expires after 1 hour of inactivity → user must re-login
- **No Database Storage**: Stateless tokens, no token table needed

## Token Storage: localStorage

### Why localStorage?

✅ **Simple** - Works seamlessly with OAuth redirect flow
✅ **Persistent** - Users stay logged in across browser sessions
✅ **Good UX** - No need to re-login on page refresh
✅ **Adequate Security** - With 1-hour expiration and XSS protection

### Implementation

```typescript
// Zustand automatically persists to localStorage
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      tokenExpiresAt: null,
      isAuthenticated: false,

      setAuth: (accessToken, user) => {
        const decoded = jwtDecode(accessToken)
        set({
          accessToken,
          user,
          tokenExpiresAt: decoded.exp * 1000,
          isAuthenticated: true
        })
      },

      clearAuth: () => set({
        user: null,
        accessToken: null,
        tokenExpiresAt: null,
        isAuthenticated: false
      })
    }),
    {
      name: 'opengov-auth',  // localStorage key
      storage: createJSONStorage(() => localStorage)
    }
  )
)
```

### Security Considerations

**XSS Mitigation:**
- Short token expiration (1 hour limits exposure)
- Content Security Policy (CSP) headers
- Input sanitization to prevent XSS attacks
- HTTPS only in production

**Token Format:**
```javascript
// Stored in localStorage as:
{
  "state": {
    "user": { "id": 123, "email": "user@example.com", ... },
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "tokenExpiresAt": 1699999999000,
    "isAuthenticated": true
  }
}
```

## Token Architecture

### JWT Structure

- **Type**: Single access token (no refresh tokens)
- **Expiration**: 1 hour
- **Algorithm**: HS256 (HMAC with SHA-256)
- **Claims**:
  - `sub`: User ID
  - `email`: User email address
  - `exp`: Expiration timestamp
  - `iat`: Issued at timestamp

### Token Lifecycle

1. **Login via Google**: Backend issues 1-hour JWT
2. **Storage**: Frontend stores in localStorage via Zustand
3. **API Calls**: Token sent in `Authorization: Bearer <token>` header
4. **Auto-Renewal**: When token is 50% expired (30 min), frontend calls `/api/auth/renew`
5. **Expiration**: After 1 hour without renewal, user must re-login

### Automatic Token Renewal

```typescript
// API client automatically renews expiring tokens
apiClient.interceptors.request.use(async (config) => {
  const { accessToken, isTokenExpiringSoon, setAuth, clearAuth } = useAuthStore.getState()

  if (!accessToken) return config

  // Renew if token expires in <10 minutes
  if (isTokenExpiringSoon()) {
    try {
      const response = await axios.post('/api/auth/renew', {}, {
        headers: { Authorization: `Bearer ${accessToken}` }
      })

      const newToken = response.data.access_token
      const user = await fetchCurrentUser(newToken)
      setAuth(newToken, user)

      config.headers.Authorization = `Bearer ${newToken}`
    } catch (error) {
      clearAuth()
      window.location.href = '/login'
    }
  } else {
    config.headers.Authorization = `Bearer ${accessToken}`
  }

  return config
})
```

## Backend Implementation

### Environment Variables

```bash
# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8000/api/auth/google/callback

# JWT
JWT_SECRET_KEY=your-random-secret-key-min-32-chars
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=60

# Frontend
FRONTEND_URL=http://localhost:5173
```

### Dependencies

- `authlib==1.3.0` - OAuth client library
- `python-jose[cryptography]==3.3.0` - JWT encoding/decoding
- `python-multipart==0.0.6` - Form data parsing

### Database Schema

**Users Table** (only table needed - no refresh tokens!)

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

### API Endpoints

- `GET /api/auth/google/login` - Initiate Google OAuth
- `GET /api/auth/google/callback` - Handle OAuth callback
- `POST /api/auth/renew` - Renew token (requires valid token)
- `POST /api/auth/logout` - Clear client-side token
- `GET /api/auth/me` - Get current user info

### Token Generation

```python
def create_access_token(data: dict) -> str:
    """Create JWT access token with 1-hour expiration"""
    to_encode = data.copy()
    expire = datetime.now(timezone.utc) + timedelta(hours=1)
    to_encode.update({
        "exp": expire,
        "iat": datetime.now(timezone.utc)
    })
    return jwt.encode(to_encode, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM)

def verify_access_token(token: str) -> Optional[dict]:
    """Verify and decode JWT access token"""
    try:
        payload = jwt.decode(token, settings.JWT_SECRET_KEY, algorithms=[settings.JWT_ALGORITHM])
        return payload
    except JWTError:
        return None
```

### Token Renewal Endpoint

```python
@router.post("/renew", response_model=TokenResponse)
async def renew_token(
    request: Request,
    current_user: User = Depends(get_current_user)
):
    """Renew access token (requires valid token)"""
    access_token = create_access_token({
        "sub": str(current_user.id),
        "email": current_user.email
    })

    return TokenResponse(
        access_token=access_token,
        token_type="bearer",
        expires_in=3600  # 1 hour
    )
```

## Security Headers

```python
# Add to backend/app/main.py
@app.middleware("http")
async def add_security_headers(request: Request, call_next):
    response = await call_next(request)

    # Prevent XSS
    response.headers["Content-Security-Policy"] = (
        "default-src 'self'; "
        "script-src 'self' 'unsafe-inline' 'unsafe-eval'; "
        "style-src 'self' 'unsafe-inline';"
    )

    # Prevent clickjacking
    response.headers["X-Frame-Options"] = "DENY"

    # XSS protection
    response.headers["X-Content-Type-Options"] = "nosniff"

    return response
```

## Google Cloud Console Setup

1. Go to https://console.cloud.google.com/
2. Create project: "OpenGov"
3. Navigate to "APIs & Services" > "Credentials"
4. Click "Create Credentials" > "OAuth client ID"
5. Select "Web application"
6. Add authorized redirect URIs:
   - Development: `http://localhost:8000/api/auth/google/callback`
   - Production: `https://yourdomain.com/api/auth/google/callback`
7. Copy Client ID and Client Secret to `.env`

## Trade-offs

### Advantages

✅ **Simple** - No database token storage, no refresh token logic
✅ **Stateless** - True stateless authentication
✅ **Fast** - No database lookups for token validation
✅ **Good UX** - Users stay logged in, auto-renewal keeps sessions alive
✅ **Scalable** - Easy to scale horizontally

### Disadvantages

❌ **Cannot Revoke** - Can't invalidate tokens before expiration
❌ **XSS Vulnerability** - localStorage accessible to JavaScript
❌ **Limited Control** - No multi-device logout

### Mitigations

- Short expiration (1 hour) limits stolen token exposure
- CSP headers prevent XSS attacks
- Can add token blacklist in Phase 2+ if needed

## Future Enhancements

If needed, we can add:
- Token blacklist for immediate revocation
- httpOnly cookies for enhanced security
- Refresh tokens for longer sessions
- Multi-device session management

## Implementation Checklist

**Backend:**
- [ ] Add dependencies (authlib, python-jose)
- [ ] Create User model
- [ ] Create database migration
- [ ] Update config.py with OAuth/JWT settings
- [ ] Implement auth service (JWT utilities)
- [ ] Create Google OAuth service
- [ ] Create auth dependencies (get_current_user)
- [ ] Implement auth router endpoints
- [ ] Add security headers middleware
- [ ] Register auth router in main.py

**Frontend:**
- [ ] Create auth store with localStorage persistence
- [ ] Update API client with auto-renewal interceptor
- [ ] Create auth callback page
- [ ] Create Google login button component
- [ ] Add protected route wrapper
- [ ] Add token renewal background hook

**Testing:**
- [ ] Test Google OAuth flow
- [ ] Test token renewal
- [ ] Test token expiration
- [ ] Test logout

**Deployment:**
- [ ] Set up Google Cloud OAuth credentials
- [ ] Configure production environment variables
- [ ] Enable HTTPS in production
