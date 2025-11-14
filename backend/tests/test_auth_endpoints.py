"""Tests for authentication API endpoints"""
import pytest
from unittest.mock import patch, AsyncMock, Mock
from fastapi.testclient import TestClient
from datetime import datetime

from app.main import app
from app.models.user import User
from app.services.auth import create_access_token
from app.routers.common import get_db


client = TestClient(app)


# Mock database dependency
def override_get_db():
    """Override database dependency for testing"""
    db = Mock()
    try:
        yield db
    finally:
        pass


app.dependency_overrides[get_db] = override_get_db


@pytest.fixture
def sample_user():
    """Create a sample user for testing"""
    return User(
        id=1,
        email="test@example.com",
        google_id="google123",
        name="Test User",
        picture_url="https://example.com/pic.jpg",
        is_active=True,
        is_verified=True,
        created_at=datetime.utcnow(),
        updated_at=datetime.utcnow(),
        last_login_at=datetime.utcnow(),
    )


@pytest.fixture
def auth_headers(sample_user):
    """Create authorization headers with valid token"""
    token = create_access_token({"sub": sample_user.id})
    return {"Authorization": f"Bearer {token}"}


class TestGoogleLoginEndpoint:
    """Test Google OAuth login endpoint"""

    def test_google_login_redirect(self):
        """Test that /api/auth/google/login redirects to Google"""
        response = client.get("/api/auth/google/login", follow_redirects=False)

        assert response.status_code == 307  # Redirect
        assert "location" in response.headers
        assert "accounts.google.com" in response.headers["location"]
        assert "client_id" in response.headers["location"]
        assert "redirect_uri" in response.headers["location"]

    @patch("app.services.auth.settings")
    def test_google_login_without_config(self, mock_settings):
        """Test login endpoint when OAuth is not configured"""
        mock_settings.GOOGLE_CLIENT_ID = ""
        mock_settings.GOOGLE_CLIENT_SECRET = ""

        response = client.get("/api/auth/google/login")

        assert response.status_code == 500
        assert "not configured" in response.json()["detail"].lower()


class TestGoogleCallbackEndpoint:
    """Test Google OAuth callback endpoint"""

    @patch("app.services.auth.exchange_code_for_user_info")
    def test_callback_success(self, mock_exchange):
        """Test successful OAuth callback"""
        # Mock the OAuth exchange
        mock_exchange.return_value = AsyncMock(return_value={
            "google_id": "google123",
            "email": "test@example.com",
            "name": "Test User",
            "picture_url": "https://example.com/pic.jpg",
        })

        # Mock database
        with patch("app.routers.auth.create_or_update_user") as mock_create_user:
            mock_user = Mock()
            mock_user.id = 1
            mock_user.email = "test@example.com"
            mock_create_user.return_value = mock_user

            response = client.get(
                "/api/auth/google/callback?code=test_auth_code",
                follow_redirects=False,
            )

            assert response.status_code == 307  # Redirect
            assert "location" in response.headers
            assert "/auth/callback" in response.headers["location"]
            assert "access_token" in response.headers["location"]

    def test_callback_without_code(self):
        """Test callback endpoint without authorization code"""
        response = client.get("/api/auth/google/callback")

        assert response.status_code == 422  # Validation error

    @patch("app.services.auth.exchange_code_for_user_info")
    def test_callback_exchange_failure(self, mock_exchange):
        """Test callback when code exchange fails"""
        from fastapi import HTTPException

        mock_exchange.side_effect = HTTPException(
            status_code=400, detail="Failed to exchange authorization code"
        )

        response = client.get(
            "/api/auth/google/callback?code=invalid_code",
            follow_redirects=False,
        )

        # Should redirect to error page
        assert response.status_code == 307
        assert "/auth/error" in response.headers["location"]


class TestAuthMeEndpoint:
    """Test /api/auth/me endpoint"""

    def test_get_current_user_success(self, sample_user, auth_headers):
        """Test getting current user info with valid token"""
        with patch("app.routers.auth.get_user_from_token") as mock_get_user:
            mock_get_user.return_value = sample_user

            response = client.get("/api/auth/me", headers=auth_headers)

            assert response.status_code == 200
            data = response.json()
            assert data["id"] == sample_user.id
            assert data["email"] == sample_user.email
            assert data["name"] == sample_user.name

    def test_get_current_user_without_token(self):
        """Test getting current user without authorization header"""
        response = client.get("/api/auth/me")

        assert response.status_code == 401
        assert "Missing authorization header" in response.json()["detail"]

    def test_get_current_user_with_invalid_token(self):
        """Test getting current user with invalid token"""
        headers = {"Authorization": "Bearer invalid_token"}
        response = client.get("/api/auth/me", headers=headers)

        assert response.status_code == 401

    def test_get_current_user_with_malformed_header(self):
        """Test getting current user with malformed authorization header"""
        headers = {"Authorization": "InvalidFormat"}
        response = client.get("/api/auth/me", headers=headers)

        assert response.status_code == 401
        assert "Invalid authorization header format" in response.json()["detail"]


class TestTokenRenewalEndpoint:
    """Test /api/auth/renew endpoint"""

    def test_renew_token_success(self, sample_user, auth_headers):
        """Test successful token renewal"""
        with patch("app.routers.auth.renew_token") as mock_renew:
            new_token = create_access_token({"sub": sample_user.id})
            mock_renew.return_value = new_token

            response = client.post("/api/auth/renew", headers=auth_headers)

            assert response.status_code == 200
            data = response.json()
            assert "access_token" in data
            assert data["token_type"] == "bearer"
            assert "expires_in" in data
            assert data["expires_in"] == 3600  # 60 minutes * 60 seconds

    def test_renew_token_without_auth(self):
        """Test token renewal without authorization header"""
        response = client.post("/api/auth/renew")

        assert response.status_code == 401
        assert "Missing authorization header" in response.json()["detail"]

    def test_renew_token_with_expired_token(self):
        """Test token renewal with expired token"""
        from datetime import timedelta
        from app.services.auth import create_access_token

        expired_token = create_access_token(
            {"sub": 1}, expires_delta=timedelta(minutes=-1)
        )
        headers = {"Authorization": f"Bearer {expired_token}"}

        response = client.post("/api/auth/renew", headers=headers)

        assert response.status_code == 401

    def test_renew_token_with_invalid_format(self):
        """Test token renewal with invalid header format"""
        headers = {"Authorization": "InvalidFormat"}
        response = client.post("/api/auth/renew", headers=headers)

        assert response.status_code == 401
        assert "Invalid authorization header format" in response.json()["detail"]


class TestLogoutEndpoint:
    """Test /api/auth/logout endpoint"""

    def test_logout_success(self):
        """Test logout endpoint returns success"""
        response = client.post("/api/auth/logout")

        assert response.status_code == 200
        data = response.json()
        assert "message" in data
        assert "client-side" in data["message"].lower()


class TestAuthenticationIntegration:
    """Integration tests for authentication flow"""

    def test_protected_endpoint_without_auth(self):
        """Test accessing protected endpoint without authentication"""
        # Try to access a protected endpoint (e.g., /api/auth/me)
        response = client.get("/api/auth/me")

        assert response.status_code == 401

    def test_full_auth_flow(self, sample_user):
        """Test complete authentication flow"""
        # 1. Create a token (simulating successful OAuth)
        token = create_access_token({"sub": sample_user.id})

        # 2. Use token to access protected endpoint
        headers = {"Authorization": f"Bearer {token}"}

        with patch("app.routers.auth.get_user_from_token") as mock_get_user:
            mock_get_user.return_value = sample_user

            # Access /me endpoint
            response = client.get("/api/auth/me", headers=headers)
            assert response.status_code == 200
            assert response.json()["email"] == sample_user.email

            # Renew token
            with patch("app.routers.auth.renew_token") as mock_renew:
                new_token = create_access_token({"sub": sample_user.id})
                mock_renew.return_value = new_token

                response = client.post("/api/auth/renew", headers=headers)
                assert response.status_code == 200
                assert "access_token" in response.json()

    def test_token_expiration_handling(self):
        """Test that expired tokens are properly rejected"""
        from datetime import timedelta

        # Create an expired token
        expired_token = create_access_token(
            {"sub": 1}, expires_delta=timedelta(minutes=-1)
        )
        headers = {"Authorization": f"Bearer {expired_token}"}

        # Try to access protected endpoint
        response = client.get("/api/auth/me", headers=headers)

        assert response.status_code == 401


class TestAuthorizationHeaderValidation:
    """Test authorization header validation"""

    def test_missing_bearer_prefix(self):
        """Test header without 'Bearer' prefix"""
        headers = {"Authorization": "some_token"}
        response = client.get("/api/auth/me", headers=headers)

        assert response.status_code == 401
        assert "Invalid authorization header format" in response.json()["detail"]

    def test_empty_token(self):
        """Test header with empty token"""
        headers = {"Authorization": "Bearer "}
        response = client.get("/api/auth/me", headers=headers)

        assert response.status_code == 401

    def test_multiple_bearer_tokens(self):
        """Test header with multiple Bearer tokens"""
        headers = {"Authorization": "Bearer token1 Bearer token2"}
        response = client.get("/api/auth/me", headers=headers)

        assert response.status_code == 401
