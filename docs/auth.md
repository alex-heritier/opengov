# Authentication System

## Overview

OpenGov uses a simple JWT (JSON Web Token) based authentication system for user sessions.

## Design Principles

- **Simplicity**: Single token type, no refresh token complexity
- **Security**: Secure token generation with reasonable expiration
- **Stateless**: No server-side session storage required

## Token Architecture

### JWT Token Structure

- **Type**: Single access token (no refresh tokens)
- **Expiration**: 24 hours
- **Algorithm**: HS256 (HMAC with SHA-256)
- **Claims**:
  - `sub`: User ID
  - `email`: User email address
  - `exp`: Expiration timestamp
  - `iat`: Issued at timestamp

### Token Lifecycle

1. **Login**: User provides credentials (email + password)
2. **Token Generation**: Server validates credentials and issues JWT token valid for 24 hours
3. **Authorization**: Client includes token in `Authorization: Bearer <token>` header
4. **Validation**: Server validates token signature and expiration on each request
5. **Expiration**: After 24 hours, user must login again to obtain new token

## Implementation Details

### Token Generation
```python
# Encode user data into JWT with 24h expiration
token = jwt.encode(
    {
        "sub": user.id,
        "email": user.email,
        "exp": datetime.utcnow() + timedelta(hours=24),
        "iat": datetime.utcnow()
    },
    SECRET_KEY,
    algorithm="HS256"
)
```

### Token Validation
```python
# Decode and validate token
payload = jwt.decode(token, SECRET_KEY, algorithms=["HS256"])
# jwt.decode() automatically validates expiration
```

## Security Considerations

- **Secret Key**: Strong random secret key stored in environment variables
- **HTTPS Only**: Tokens transmitted only over HTTPS in production
- **Token Storage**: Client stores token in httpOnly cookies or secure localStorage
- **No Sensitive Data**: Tokens contain only user ID and email (no passwords or sensitive data)

## Trade-offs

### Advantages
- Simple implementation and maintenance
- No database lookups for token refresh
- Stateless authentication
- Easy to scale horizontally

### Disadvantages
- Cannot revoke tokens before expiration (use short expiration as mitigation)
- Users must re-login every 24 hours
- No "remember me" functionality without extending expiration

## Future Enhancements

If needed, we can add:
- Refresh tokens for longer sessions
- Token blacklisting for immediate revocation
- Sliding expiration windows
- Multiple device session management
