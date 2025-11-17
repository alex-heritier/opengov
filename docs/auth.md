# Authentication System Documentation

## Overview

The OpenGov API uses **fastapi-users** with **cookie-based authentication** for secure, session-like user management. This provides a Laravel-style authentication experience with HTTP-only secure cookies.

This is the **current production authentication system** (as of Nov 2025).

## Architecture

- **Library**: `fastapi-users` v12.x (currently implemented)
- **Transport**: Cookie-based (HTTP-only, secure)
- **Strategy**: JWT tokens stored in cookies
- **Password Hashing**: bcrypt via passlib
- **Database**: Async SQLAlchemy with SQLite (aiosqlite)
- **Future**: Google OAuth 2.0 fields reserved in User model for Phase 2

## Authentication Flow

1. **Registration**: User creates account with email/password
2. **Login**: Credentials validated, JWT token set in HTTP-only cookie
3. **Authenticated Requests**: Cookie automatically sent with each request
4. **Logout**: Cookie cleared from client

## API Endpoints

### Public Endpoints

#### Register
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "John Doe" (optional)
}
```

**Response** (201 Created):
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "is_active": true,
  "is_superuser": false,
  "is_verified": false,
  "created_at": "2025-11-17T04:27:59.488665",
  "updated_at": "2025-11-17T04:27:59.488669"
}
```

#### Login
```http
POST /api/auth/login
Content-Type: application/x-www-form-urlencoded

username=user@example.com&password=securepassword123
```

**Response** (204 No Content):
- Sets `opengov_auth` cookie with JWT token
- Cookie is HTTP-only, secure (in production), SameSite=lax

#### Logout
```http
POST /api/auth/logout
```

**Response** (204 No Content):
- Clears authentication cookie

#### Forgot Password
```http
POST /api/auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

### Protected Endpoints

All protected endpoints require a valid authentication cookie.

#### Get Current User
```http
GET /api/users/me
```

**Response** (200 OK):
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "picture_url": "https://example.com/avatar.jpg",
  "google_id": null,
  "is_active": true,
  "is_superuser": false,
  "is_verified": false,
  "created_at": "2025-11-17T04:27:59.488665",
  "updated_at": "2025-11-17T04:27:59.488669",
  "last_login_at": "2025-11-17T10:30:00.000000"
}
```

#### Update Current User
```http
PATCH /api/users/me
Content-Type: application/json

{
  "name": "Jane Doe",
  "picture_url": "https://example.com/avatar.jpg"
}
```

**Response** (200 OK):
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "Jane Doe",
  "picture_url": "https://example.com/avatar.jpg",
  "google_id": null,
  "is_active": true,
  "is_superuser": false,
  "is_verified": false,
  "created_at": "2025-11-17T04:27:59.488665",
  "updated_at": "2025-11-17T10:35:00.000000",
  "last_login_at": "2025-11-17T10:30:00.000000"
}
```

## Security Features

### Cookie Configuration

The authentication cookie (`opengov_auth`) has the following security attributes:

- **HttpOnly**: `true` - Prevents JavaScript access (XSS protection)
- **Secure**: Configurable via `COOKIE_SECURE` environment variable
  - `false` in development (HTTP)
  - `true` in production (HTTPS required)
- **SameSite**: `lax` - CSRF protection
- **Max-Age**: Configurable via `JWT_ACCESS_TOKEN_EXPIRE_MINUTES` (default: 60 minutes)

### Password Security

- **Hashing Algorithm**: bcrypt (work factor automatically managed by passlib)
- **Minimum Password Length**: Enforced by fastapi-users (8 characters)
- **Password Reset**: Secure token-based flow

### JWT Token Security

- **Secret Key**: Configured via `JWT_SECRET_KEY` environment variable
- **Algorithm**: HS256
- **Lifetime**: Configurable via `JWT_ACCESS_TOKEN_EXPIRE_MINUTES`
- **Audience**: `fastapi-users:auth`

## Environment Configuration

### Required Variables

```bash
# JWT Configuration
JWT_SECRET_KEY=<minimum 32 characters, use secrets.token_urlsafe(32)>
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=60

# Database (async support required)
DATABASE_URL=sqlite:///./opengov.db

# Authentication Security
COOKIE_SECURE=false  # Set to 'true' in production with HTTPS
```

### Production Checklist

- [ ] Set `COOKIE_SECURE=true`
- [ ] Use HTTPS/TLS for all connections
- [ ] Generate strong `JWT_SECRET_KEY` (32+ characters)
- [ ] Configure proper CORS origins
- [ ] Set `ENVIRONMENT=production`
- [ ] Enable `BEHIND_PROXY=true` if using reverse proxy
- [ ] Configure email sending for password reset (future)

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    hashed_password VARCHAR(1024) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    is_superuser BOOLEAN DEFAULT FALSE NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE NOT NULL,

    -- Optional OAuth fields (for future Google OAuth)
    google_id VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    picture_url VARCHAR(500),

    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_login_at DATETIME
);
```

## Migration from Old Auth System

If upgrading from the previous manual JWT/OAuth system:

1. **Backup database**: `cp opengov.db opengov.db.backup`
2. **Run migrations**: `alembic upgrade head`
3. **Existing users**: Will need to use password reset flow to set passwords
4. **OAuth users**: Google OAuth fields preserved for Phase 2

## Development

### Testing Authentication

```bash
# Register a user
curl -X POST http://localhost:8000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass123"}'

# Login (saves cookie)
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=test@example.com&password=testpass123" \
  -c cookies.txt

# Access protected endpoint
curl -X GET http://localhost:8000/api/users/me \
  -b cookies.txt

# Logout
curl -X POST http://localhost:8000/api/auth/logout \
  -b cookies.txt \
  -c cookies.txt
```

### Custom User Manager Events

The `UserManager` class in `app/auth.py` provides hooks for custom logic:

- `on_after_register`: Called after successful registration
- `on_after_forgot_password`: Called when password reset requested
- `on_after_request_verify`: Called when email verification requested

## Common Issues

### Cookie Not Being Set

**Cause**: CORS configuration blocking credentials

**Solution**: Ensure frontend sends `credentials: 'include'` and backend has proper CORS config:
```python
allow_credentials=True,
allow_origins=["http://localhost:5173"]  # Your frontend URL
```

### 401 Unauthorized on Protected Endpoints

**Cause**: Cookie not being sent or expired

**Solutions**:
1. Check cookie exists in browser DevTools
2. Verify cookie domain matches request domain
3. Ensure cookie hasn't expired
4. Check `COOKIE_SECURE` matches protocol (HTTP vs HTTPS)

### bcrypt Warning

**Issue**: Warning about `pkg_resources` deprecation

**Impact**: Non-critical, doesn't affect functionality

**Fix**: This is a known passlib/bcrypt compatibility issue. It's handled by pinning bcrypt to v4.x.

## Future Enhancements (Phase 2)

- [ ] Google OAuth 2.0 integration (OAuth fields already in User schema)
- [ ] Email verification workflow
- [ ] Password reset email notifications
- [ ] Two-factor authentication
- [ ] Session management (revocation)
- [ ] Remember me functionality
- [ ] Account deletion workflow

**Note:** Google OAuth fields (`google_id`, `picture_url`, `name`) are reserved in the User model for Phase 2 implementation. See `docs/google_oauth_plan.md` for detailed implementation plan (currently Phase 2 - not yet implemented).

## References

- [fastapi-users Documentation](https://fastapi-users.github.io/fastapi-users/)
- [Cookie Authentication Transport](https://fastapi-users.github.io/fastapi-users/configuration/authentication/transports/#cookie)
- [JWT Strategy](https://fastapi-users.github.io/fastapi-users/configuration/authentication/strategies/jwt/)
