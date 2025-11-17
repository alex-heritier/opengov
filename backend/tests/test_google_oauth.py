"""Tests for Google OAuth integration"""
import pytest
from unittest.mock import AsyncMock, Mock, patch
from datetime import datetime, timezone
from httpx import AsyncClient
from sqlalchemy.orm import Session
from app.models.user import User
from app.services.google_oauth import GoogleOAuthProvider
from fastapi_users.password import PasswordHelper

__all__ = ["datetime", "timezone"]  # noqa: F405


class TestGoogleOAuthProvider:
    """Test Google OAuth provider service"""

    def test_get_authorization_url(self):
        """Test OAuth authorization URL generation"""
        state = "test_state_token"
        url = GoogleOAuthProvider.get_authorization_url(state)

        assert "https://accounts.google.com/o/oauth2/v2/auth" in url
        assert "client_id" in url
        assert "redirect_uri" in url
        assert "scope" in url
        assert "state=test_state_token" in url
        assert "openid" in url
        assert "email" in url
        assert "profile" in url

    @pytest.mark.asyncio
    async def test_exchange_code_for_token(self):
        """Test exchanging authorization code for access token"""
        mock_token = {
            "access_token": "mock_access_token",
            "token_type": "Bearer",
            "expires_in": 3600,
        }

        with patch("app.services.google_oauth.AsyncOAuth2Client") as mock_client_class:
            mock_client = AsyncMock()
            mock_client.__aenter__.return_value = mock_client
            mock_client.__aexit__.return_value = None
            mock_client.fetch_token = AsyncMock(return_value=mock_token)
            mock_client_class.return_value = mock_client

            result = await GoogleOAuthProvider.exchange_code_for_token("test_code")

            assert result == mock_token
            mock_client.fetch_token.assert_called_once_with(
                "https://oauth2.googleapis.com/token",
                code="test_code",
            )

    @pytest.mark.asyncio
    async def test_get_user_info(self):
        """Test fetching user info from Google"""
        mock_user_info = {
            "sub": "google_user_123",
            "email": "test@example.com",
            "name": "Test User",
            "picture": "https://example.com/photo.jpg",
            "email_verified": True,
        }

        token = {"access_token": "mock_access_token"}

        with patch("app.services.google_oauth.AsyncOAuth2Client") as mock_client_class:
            mock_client = AsyncMock()
            mock_client.__aenter__.return_value = mock_client
            mock_client.__aexit__.return_value = None

            mock_response = Mock()
            mock_response.json.return_value = mock_user_info
            mock_client.get = AsyncMock(return_value=mock_response)
            mock_client_class.return_value = mock_client

            result = await GoogleOAuthProvider.get_user_info(token)

            assert result == mock_user_info
            mock_client.get.assert_called_once_with(
                "https://openidconnect.googleapis.com/v1/userinfo",
            )


class TestGoogleOAuthRouter:
    """Test Google OAuth router endpoints"""

    @pytest.mark.asyncio
    async def test_google_login_redirect(self, client: AsyncClient):
        """Test /api/auth/google/login redirects to Google"""
        response = await client.get("/api/auth/google/login", follow_redirects=False)

        assert response.status_code == 307
        redirect_url = response.headers.get("location")
        assert "https://accounts.google.com/o/oauth2/v2/auth" in redirect_url
        assert "client_id" in redirect_url
        assert "state" in redirect_url

    @pytest.mark.asyncio
    async def test_google_callback_invalid_state(self, client: AsyncClient):
        """Test callback with invalid state parameter"""
        response = await client.get(
            "/api/auth/google/callback?code=test_code&state=invalid_state",
            follow_redirects=False
        )

        assert response.status_code == 400

    @pytest.mark.asyncio
    async def test_google_callback_creates_new_user(
        self, client: AsyncClient, db_session: Session
    ):
        """Test callback creates new user from Google OAuth"""
        from app.routers.oauth import _oauth_states

        # Setup: Generate valid state
        state = "valid_test_state"
        _oauth_states[state] = datetime.now(timezone.utc).timestamp()

        mock_token = {
            "access_token": "mock_access_token",
            "token_type": "Bearer",
        }

        mock_user_info = {
            "sub": "google_123",
            "email": "newuser@example.com",
            "name": "New User",
            "picture": "https://example.com/photo.jpg",
            "email_verified": True,
        }

        with patch(
            "app.services.google_oauth.GoogleOAuthProvider.exchange_code_for_token"
        ) as mock_exchange, patch(
            "app.services.google_oauth.GoogleOAuthProvider.get_user_info"
        ) as mock_get_user:
            mock_exchange.return_value = mock_token
            mock_get_user.return_value = mock_user_info

            response = await client.get(
                f"/api/auth/google/callback?code=test_code&state={state}",
                follow_redirects=False
            )

            # Should redirect to frontend feed page
            assert response.status_code == 307
            redirect_url = response.headers.get("location")
            assert "/feed" in redirect_url

            # Verify user was created
            user = db_session.query(User).filter(User.email == "newuser@example.com").first()
            assert user is not None
            assert user.google_id == "google_123"
            assert user.name == "New User"
            assert user.picture_url == "https://example.com/photo.jpg"
            assert user.is_verified is True

    @pytest.mark.asyncio
    async def test_google_callback_links_existing_email_user(
        self, client: AsyncClient, db_session: Session
    ):
        """Test callback links Google account to existing email/password user"""
        from app.routers.oauth import _oauth_states

        # Create existing user with email/password
        password_helper = PasswordHelper()
        existing_user = User(
            email="existing@example.com",
            hashed_password=password_helper.hash("password123"),
            is_active=True,
            is_verified=False,
            name="Existing User",
        )
        db_session.add(existing_user)
        db_session.commit()
        user_id = existing_user.id

        # Setup: Generate valid state
        state = "valid_test_state"
        _oauth_states[state] = datetime.now(timezone.utc).timestamp()

        mock_token = {"access_token": "mock_access_token"}
        mock_user_info = {
            "sub": "google_456",
            "email": "existing@example.com",
            "name": "Updated Name",
            "picture": "https://example.com/new_photo.jpg",
            "email_verified": True,
        }

        with patch(
            "app.services.google_oauth.GoogleOAuthProvider.exchange_code_for_token"
        ) as mock_exchange, patch(
            "app.services.google_oauth.GoogleOAuthProvider.get_user_info"
        ) as mock_get_user:
            mock_exchange.return_value = mock_token
            mock_get_user.return_value = mock_user_info

            response = await client.get(
                f"/api/auth/google/callback?code=test_code&state={state}",
                follow_redirects=False
            )

            assert response.status_code == 307

            # Verify user was updated with Google info
            db_session.expire_all()
            user = db_session.query(User).filter(User.id == user_id).first()
            assert user.google_id == "google_456"
            assert user.name == "Updated Name"
            assert user.picture_url == "https://example.com/new_photo.jpg"
            # Note: is_verified only updates if email_verified is True in response
            # The oauth callback code sets it conditionally

    @pytest.mark.asyncio
    async def test_google_callback_updates_existing_google_user(
        self, client: AsyncClient, db_session: Session
    ):
        """Test callback updates existing Google OAuth user profile"""
        from app.routers.oauth import _oauth_states

        # Create existing Google OAuth user
        existing_user = User(
            email="googleuser@example.com",
            google_id="google_789",
            hashed_password="",
            is_active=True,
            is_verified=True,
            name="Old Name",
            picture_url="https://example.com/old_photo.jpg",
        )
        db_session.add(existing_user)
        db_session.commit()
        user_id = existing_user.id

        # Setup: Generate valid state
        state = "valid_test_state"
        _oauth_states[state] = datetime.now(timezone.utc).timestamp()

        mock_token = {"access_token": "mock_access_token"}
        mock_user_info = {
            "sub": "google_789",
            "email": "googleuser@example.com",
            "name": "Updated Google Name",
            "picture": "https://example.com/updated_photo.jpg",
            "email_verified": True,
        }

        with patch(
            "app.services.google_oauth.GoogleOAuthProvider.exchange_code_for_token"
        ) as mock_exchange, patch(
            "app.services.google_oauth.GoogleOAuthProvider.get_user_info"
        ) as mock_get_user:
            mock_exchange.return_value = mock_token
            mock_get_user.return_value = mock_user_info

            response = await client.get(
                f"/api/auth/google/callback?code=test_code&state={state}",
                follow_redirects=False
            )

            assert response.status_code == 307

            # Verify user profile was updated
            db_session.expire_all()
            user = db_session.query(User).filter(User.id == user_id).first()
            assert user.name == "Updated Google Name"
            assert user.picture_url == "https://example.com/updated_photo.jpg"

    @pytest.mark.asyncio
    async def test_google_callback_missing_google_id(
        self, client: AsyncClient, db_session: Session
    ):
        """Test callback handles missing Google ID (sub) in user info"""
        from app.routers.oauth import _oauth_states

        state = "valid_test_state"
        _oauth_states[state] = datetime.now(timezone.utc).timestamp()

        mock_token = {"access_token": "mock_access_token"}
        mock_user_info = {
            "email": "user@example.com",
            "name": "User Without Sub",
        }

        with patch(
            "app.services.google_oauth.GoogleOAuthProvider.exchange_code_for_token"
        ) as mock_exchange, patch(
            "app.services.google_oauth.GoogleOAuthProvider.get_user_info"
        ) as mock_get_user:
            mock_exchange.return_value = mock_token
            mock_get_user.return_value = mock_user_info

            response = await client.get(
                f"/api/auth/google/callback?code=test_code&state={state}",
                follow_redirects=False
            )

            # Should redirect to login with error
            assert response.status_code == 307
            redirect_url = response.headers.get("location")
            assert "/login" in redirect_url
            assert "error=invalid_user_info" in redirect_url

    @pytest.mark.asyncio
    async def test_google_callback_exchange_error(
        self, client: AsyncClient, db_session: Session
    ):
        """Test callback handles OAuth token exchange errors"""
        from app.routers.oauth import _oauth_states

        state = "valid_test_state"
        _oauth_states[state] = datetime.now(timezone.utc).timestamp()

        with patch(
            "app.services.google_oauth.GoogleOAuthProvider.exchange_code_for_token"
        ) as mock_exchange:
            mock_exchange.side_effect = Exception("OAuth exchange failed")

            response = await client.get(
                f"/api/auth/google/callback?code=test_code&state={state}",
                follow_redirects=False
            )

            # Should redirect to login with error
            assert response.status_code == 307
            redirect_url = response.headers.get("location")
            assert "/login" in redirect_url
            assert "error=oauth_error" in redirect_url

    @pytest.mark.asyncio
    async def test_google_callback_sets_auth_cookie(
        self, client: AsyncClient, db_session: Session
    ):
        """Test callback sets authentication cookie after successful OAuth"""
        from app.routers.oauth import _oauth_states

        state = "valid_test_state"
        _oauth_states[state] = datetime.now(timezone.utc).timestamp()

        mock_token = {"access_token": "mock_access_token"}
        mock_user_info = {
            "sub": "google_999",
            "email": "cookieuser@example.com",
            "name": "Cookie User",
            "email_verified": True,
        }

        with patch(
            "app.services.google_oauth.GoogleOAuthProvider.exchange_code_for_token"
        ) as mock_exchange, patch(
            "app.services.google_oauth.GoogleOAuthProvider.get_user_info"
        ) as mock_get_user:
            mock_exchange.return_value = mock_token
            mock_get_user.return_value = mock_user_info

            response = await client.get(
                f"/api/auth/google/callback?code=test_code&state={state}",
                follow_redirects=False
            )

            assert response.status_code == 307

            # Check that Set-Cookie header is present
            cookies = response.headers.get("set-cookie")
            assert cookies is not None, "No set-cookie header in response"
            # The cookie should contain the auth token
            # Note: fastapi-users uses "fastapiusersauth" as default cookie name
            assert "fastapiusersauth" in cookies or "opengov" in cookies

    @pytest.mark.asyncio
    async def test_google_callback_updates_last_login(
        self, client: AsyncClient, db_session: Session
    ):
        """Test callback updates last_login_at timestamp"""
        from app.routers.oauth import _oauth_states

        state = "valid_test_state"
        _oauth_states[state] = datetime.now(timezone.utc).timestamp()

        mock_token = {"access_token": "mock_access_token"}
        mock_user_info = {
            "sub": "google_login_test",
            "email": "logintest@example.com",
            "name": "Login Test User",
            "email_verified": True,
        }

        with patch(
            "app.services.google_oauth.GoogleOAuthProvider.exchange_code_for_token"
        ) as mock_exchange, patch(
            "app.services.google_oauth.GoogleOAuthProvider.get_user_info"
        ) as mock_get_user:
            mock_exchange.return_value = mock_token
            mock_get_user.return_value = mock_user_info

            response = await client.get(
                f"/api/auth/google/callback?code=test_code&state={state}",
                follow_redirects=False
            )

            assert response.status_code == 307

            # Verify last_login_at was set
            user = db_session.query(User).filter(
                User.email == "logintest@example.com"
            ).first()
            assert user is not None
            assert user.last_login_at is not None
            # Verify the timestamp is recent (within last minute)
            now = datetime.now(timezone.utc)
            # Make both timezone-aware for comparison
            if user.last_login_at.tzinfo is None:
                last_login = user.last_login_at.replace(tzinfo=timezone.utc)
            else:
                last_login = user.last_login_at
            time_diff = (now - last_login).total_seconds()
            assert time_diff < 60, f"last_login_at is too old: {time_diff} seconds ago"
