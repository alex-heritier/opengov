"""Google OAuth provider for fastapi-users integration"""
import logging
from authlib.integrations.httpx_client import AsyncOAuth2Client
from app.config import settings

logger = logging.getLogger(__name__)

# Required scopes for user info
REQUIRED_SCOPES = ["openid", "email", "profile"]


class GoogleOAuthProvider:
    """Helper for Google OAuth with fastapi-users"""

    @staticmethod
    def get_authorization_url(state: str) -> str:
        """Generate Google OAuth authorization URL"""
        client = AsyncOAuth2Client(
            client_id=settings.GOOGLE_CLIENT_ID,
            redirect_uri=settings.GOOGLE_REDIRECT_URI,
        )
        auth_url, auth_state = client.create_authorization_url(
            "https://accounts.google.com/o/oauth2/v2/auth",
            scope=REQUIRED_SCOPES,
            state=state,
        )
        return auth_url

    @staticmethod
    async def exchange_code_for_token(code: str) -> dict:
        """Exchange authorization code for access token"""
        try:
            async with AsyncOAuth2Client(
                client_id=settings.GOOGLE_CLIENT_ID,
                client_secret=settings.GOOGLE_CLIENT_SECRET.get_secret_value(),
                redirect_uri=settings.GOOGLE_REDIRECT_URI,
            ) as client:
                token = await client.fetch_token(
                    "https://oauth2.googleapis.com/token",
                    code=code,
                )
                return token
        except Exception as e:
            logger.error(f"Failed to exchange code for token: {e}")
            raise RuntimeError(f"Token exchange failed: {e}") from e

    @staticmethod
    async def get_user_info(token: dict) -> dict:
        """Fetch user info from Google"""
        if not token or "access_token" not in token:
            raise ValueError("Invalid token: missing access_token")

        try:
            async with AsyncOAuth2Client(
                client_id=settings.GOOGLE_CLIENT_ID,
                token=token,
            ) as client:
                response = await client.get(
                    "https://openidconnect.googleapis.com/v1/userinfo",
                )
                response.raise_for_status()
                return response.json()
        except Exception as e:
            logger.error(f"Failed to fetch user info: {e}")
            raise ValueError(f"Failed to fetch user info: {e}") from e
