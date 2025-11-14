"""Comprehensive tests for authentication system"""
import pytest
from datetime import datetime, timedelta
from unittest.mock import Mock, patch, AsyncMock
from fastapi import HTTPException
from jose import jwt

from app.services.auth import (
    create_access_token,
    decode_access_token,
    get_user_from_token,
    create_or_update_user,
    renew_token,
)
from app.models.user import User
from app.config import settings


# Test fixtures
@pytest.fixture
def mock_db_session(mocker):
    """Create a mock database session"""
    return mocker.MagicMock()


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


# JWT Token Tests
class TestJWTTokens:
    """Test JWT token creation and validation"""

    def test_create_access_token(self):
        """Test creating a valid access token"""
        data = {"sub": 1}
        token = create_access_token(data)

        assert isinstance(token, str)
        assert len(token) > 0

        # Decode and verify token
        payload = jwt.decode(
            token, settings.JWT_SECRET_KEY, algorithms=[settings.JWT_ALGORITHM]
        )
        assert payload["sub"] == 1
        assert "exp" in payload
        assert "iat" in payload

    def test_create_access_token_with_custom_expiration(self):
        """Test creating token with custom expiration"""
        data = {"sub": 1}
        expires_delta = timedelta(minutes=30)
        token = create_access_token(data, expires_delta)

        payload = jwt.decode(
            token, settings.JWT_SECRET_KEY, algorithms=[settings.JWT_ALGORITHM]
        )

        # Check expiration is approximately 30 minutes from now
        exp_time = datetime.fromtimestamp(payload["exp"])
        now = datetime.utcnow()
        time_diff = (exp_time - now).total_seconds()

        assert 29 * 60 < time_diff < 31 * 60  # Allow 1 minute margin

    def test_decode_valid_token(self):
        """Test decoding a valid token"""
        data = {"sub": 1, "email": "test@example.com"}
        token = create_access_token(data)

        payload = decode_access_token(token)
        assert payload["sub"] == 1
        assert payload["email"] == "test@example.com"

    def test_decode_expired_token(self):
        """Test decoding an expired token raises exception"""
        data = {"sub": 1}
        # Create token that expired 1 minute ago
        expires_delta = timedelta(minutes=-1)
        token = create_access_token(data, expires_delta)

        with pytest.raises(HTTPException) as exc_info:
            decode_access_token(token)

        assert exc_info.value.status_code == 401
        assert "Could not validate credentials" in str(exc_info.value.detail)

    def test_decode_invalid_token(self):
        """Test decoding an invalid token raises exception"""
        invalid_token = "invalid.token.here"

        with pytest.raises(HTTPException) as exc_info:
            decode_access_token(invalid_token)

        assert exc_info.value.status_code == 401

    def test_decode_token_with_wrong_secret(self):
        """Test decoding token with wrong secret raises exception"""
        # Create token with different secret
        wrong_token = jwt.encode(
            {"sub": 1, "exp": datetime.utcnow() + timedelta(hours=1)},
            "wrong-secret-key",
            algorithm="HS256",
        )

        with pytest.raises(HTTPException) as exc_info:
            decode_access_token(wrong_token)

        assert exc_info.value.status_code == 401


# User Authentication Tests
class TestUserAuthentication:
    """Test user authentication and retrieval"""

    def test_get_user_from_valid_token(self, mock_db_session, sample_user):
        """Test getting user from valid token"""
        # Create token
        token = create_access_token({"sub": sample_user.id})

        # Mock database query
        mock_query = Mock()
        mock_query.filter.return_value.first.return_value = sample_user
        mock_db_session.query.return_value = mock_query

        # Get user
        user = get_user_from_token(token, mock_db_session)

        assert user.id == sample_user.id
        assert user.email == sample_user.email
        mock_db_session.query.assert_called_once_with(User)

    def test_get_user_from_token_user_not_found(self, mock_db_session):
        """Test getting user when user doesn't exist in database"""
        token = create_access_token({"sub": 999})

        # Mock database query to return None
        mock_query = Mock()
        mock_query.filter.return_value.first.return_value = None
        mock_db_session.query.return_value = mock_query

        with pytest.raises(HTTPException) as exc_info:
            get_user_from_token(token, mock_db_session)

        assert exc_info.value.status_code == 401
        assert "User not found" in str(exc_info.value.detail)

    def test_get_user_from_token_inactive_user(self, mock_db_session, sample_user):
        """Test getting inactive user raises exception"""
        token = create_access_token({"sub": sample_user.id})
        sample_user.is_active = False

        mock_query = Mock()
        mock_query.filter.return_value.first.return_value = sample_user
        mock_db_session.query.return_value = mock_query

        with pytest.raises(HTTPException) as exc_info:
            get_user_from_token(token, mock_db_session)

        assert exc_info.value.status_code == 403
        assert "Inactive user" in str(exc_info.value.detail)

    def test_get_user_from_token_missing_sub(self, mock_db_session):
        """Test getting user from token without 'sub' claim"""
        token = create_access_token({"email": "test@example.com"})

        with pytest.raises(HTTPException) as exc_info:
            get_user_from_token(token, mock_db_session)

        assert exc_info.value.status_code == 401


# User Creation/Update Tests
class TestUserCreationAndUpdate:
    """Test creating and updating users from OAuth data"""

    def test_create_new_user(self, mock_db_session):
        """Test creating a new user from Google OAuth data"""
        # Mock database to return no existing user
        mock_query = Mock()
        mock_query.filter.return_value.first.return_value = None
        mock_db_session.query.return_value = mock_query

        user = create_or_update_user(
            db=mock_db_session,
            google_id="google123",
            email="newuser@example.com",
            name="New User",
            picture_url="https://example.com/pic.jpg",
        )

        # Verify user was added
        mock_db_session.add.assert_called_once()
        mock_db_session.commit.assert_called_once()
        mock_db_session.refresh.assert_called_once()

    def test_update_existing_user_by_google_id(self, mock_db_session, sample_user):
        """Test updating existing user found by google_id"""
        # Mock database to return existing user
        mock_query = Mock()
        mock_query.filter.return_value.first.return_value = sample_user
        mock_db_session.query.return_value = mock_query

        user = create_or_update_user(
            db=mock_db_session,
            google_id="google123",
            email="updated@example.com",
            name="Updated Name",
            picture_url="https://example.com/newpic.jpg",
        )

        # Verify user was updated (not added)
        mock_db_session.add.assert_not_called()
        mock_db_session.commit.assert_called_once()

        # Verify fields were updated
        assert sample_user.email == "updated@example.com"
        assert sample_user.name == "Updated Name"
        assert sample_user.is_verified is True

    def test_update_existing_user_by_email(self, mock_db_session, sample_user):
        """Test updating existing user found by email"""
        sample_user.google_id = None  # User exists but no google_id yet

        mock_query = Mock()
        mock_query.filter.return_value.first.return_value = sample_user
        mock_db_session.query.return_value = mock_query

        user = create_or_update_user(
            db=mock_db_session,
            google_id="new_google_id",
            email=sample_user.email,
            name="Updated Name",
            picture_url="https://example.com/pic.jpg",
        )

        # Verify google_id was added
        assert sample_user.google_id == "new_google_id"
        assert sample_user.is_verified is True
        mock_db_session.commit.assert_called_once()


# Token Renewal Tests
class TestTokenRenewal:
    """Test token renewal functionality"""

    def test_renew_valid_token(self, mock_db_session, sample_user):
        """Test renewing a valid token"""
        # Create current token
        current_token = create_access_token({"sub": sample_user.id})

        # Mock database query
        mock_query = Mock()
        mock_query.filter.return_value.first.return_value = sample_user
        mock_db_session.query.return_value = mock_query

        # Renew token
        new_token = renew_token(current_token, mock_db_session)

        assert isinstance(new_token, str)
        assert new_token != current_token

        # Verify new token is valid
        payload = decode_access_token(new_token)
        assert payload["sub"] == sample_user.id

    def test_renew_expired_token_fails(self, mock_db_session):
        """Test renewing an expired token fails"""
        # Create expired token
        expired_token = create_access_token(
            {"sub": 1}, expires_delta=timedelta(minutes=-1)
        )

        with pytest.raises(HTTPException) as exc_info:
            renew_token(expired_token, mock_db_session)

        assert exc_info.value.status_code == 401

    def test_renew_token_inactive_user_fails(self, mock_db_session, sample_user):
        """Test renewing token for inactive user fails"""
        current_token = create_access_token({"sub": sample_user.id})
        sample_user.is_active = False

        mock_query = Mock()
        mock_query.filter.return_value.first.return_value = sample_user
        mock_db_session.query.return_value = mock_query

        with pytest.raises(HTTPException) as exc_info:
            renew_token(current_token, mock_db_session)

        assert exc_info.value.status_code == 403


# OAuth Exchange Tests
class TestOAuthExchange:
    """Test Google OAuth code exchange"""

    @pytest.mark.asyncio
    async def test_exchange_code_success(self):
        """Test successful code exchange for user info"""
        from app.services.auth import exchange_code_for_user_info

        mock_token_response = Mock()
        mock_token_response.status_code = 200
        mock_token_response.json.return_value = {"access_token": "fake_token"}

        mock_userinfo_response = Mock()
        mock_userinfo_response.status_code = 200
        mock_userinfo_response.json.return_value = {
            "id": "google123",
            "email": "test@example.com",
            "name": "Test User",
            "picture": "https://example.com/pic.jpg",
        }

        with patch("httpx.AsyncClient") as mock_client:
            mock_instance = mock_client.return_value.__aenter__.return_value
            mock_instance.post = AsyncMock(return_value=mock_token_response)
            mock_instance.get = AsyncMock(return_value=mock_userinfo_response)

            user_info = await exchange_code_for_user_info("auth_code")

            assert user_info["google_id"] == "google123"
            assert user_info["email"] == "test@example.com"
            assert user_info["name"] == "Test User"
            assert user_info["picture_url"] == "https://example.com/pic.jpg"

    @pytest.mark.asyncio
    async def test_exchange_code_token_failure(self):
        """Test code exchange when token request fails"""
        from app.services.auth import exchange_code_for_user_info

        mock_response = Mock()
        mock_response.status_code = 400
        mock_response.text = "Invalid code"

        with patch("httpx.AsyncClient") as mock_client:
            mock_instance = mock_client.return_value.__aenter__.return_value
            mock_instance.post = AsyncMock(return_value=mock_response)

            with pytest.raises(HTTPException) as exc_info:
                await exchange_code_for_user_info("invalid_code")

            assert exc_info.value.status_code == 400
            assert "Failed to exchange authorization code" in str(exc_info.value.detail)

    @pytest.mark.asyncio
    async def test_exchange_code_userinfo_failure(self):
        """Test code exchange when userinfo request fails"""
        from app.services.auth import exchange_code_for_user_info

        mock_token_response = Mock()
        mock_token_response.status_code = 200
        mock_token_response.json.return_value = {"access_token": "fake_token"}

        mock_userinfo_response = Mock()
        mock_userinfo_response.status_code = 401
        mock_userinfo_response.text = "Invalid token"

        with patch("httpx.AsyncClient") as mock_client:
            mock_instance = mock_client.return_value.__aenter__.return_value
            mock_instance.post = AsyncMock(return_value=mock_token_response)
            mock_instance.get = AsyncMock(return_value=mock_userinfo_response)

            with pytest.raises(HTTPException) as exc_info:
                await exchange_code_for_user_info("auth_code")

            assert exc_info.value.status_code == 400
            assert "Failed to fetch user information" in str(exc_info.value.detail)
