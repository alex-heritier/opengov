"""Custom exceptions for the OpenGov application"""
from typing import Optional


class OpenGovException(Exception):
    """Base exception for all OpenGov errors"""
    def __init__(
        self,
        message: str,
        code: str,
        status_code: int = 500,
        action: Optional[str] = None,
    ):
        self.message = message
        self.code = code
        self.status_code = status_code
        self.action = action
        super().__init__(message)


# Authentication Exceptions
class AuthenticationError(OpenGovException):
    """Base class for authentication errors"""
    def __init__(self, message: str, code: str, action: Optional[str] = None):
        super().__init__(message, code, status_code=401, action=action)


class TokenExpiredError(AuthenticationError):
    """Raised when JWT token has expired"""
    def __init__(self):
        super().__init__(
            message="Your session has expired. Please sign in again.",
            code="TOKEN_EXPIRED",
            action="REDIRECT_LOGIN",
        )


class InvalidTokenError(AuthenticationError):
    """Raised when JWT token is invalid or malformed"""
    def __init__(self):
        super().__init__(
            message="Your session is invalid. Please sign in again.",
            code="TOKEN_INVALID",
            action="REDIRECT_LOGIN",
        )


class MissingTokenError(AuthenticationError):
    """Raised when authorization header is missing"""
    def __init__(self):
        super().__init__(
            message="Authentication required. Please sign in.",
            code="TOKEN_MISSING",
            action="REDIRECT_LOGIN",
        )


class InvalidAuthHeaderError(AuthenticationError):
    """Raised when authorization header format is invalid"""
    def __init__(self):
        super().__init__(
            message="Invalid authentication format. Please sign in again.",
            code="AUTH_HEADER_INVALID",
            action="REDIRECT_LOGIN",
        )


class UserNotFoundError(AuthenticationError):
    """Raised when user associated with token is not found"""
    def __init__(self):
        super().__init__(
            message="User account not found. Please sign in again.",
            code="USER_NOT_FOUND",
            action="REDIRECT_LOGIN",
        )


class InactiveUserError(OpenGovException):
    """Raised when user account is inactive"""
    def __init__(self):
        super().__init__(
            message="Your account has been deactivated. Please contact support.",
            code="USER_INACTIVE",
            status_code=403,
            action="CONTACT_SUPPORT",
        )


# OAuth Exceptions
class OAuthError(OpenGovException):
    """Base class for OAuth errors"""
    def __init__(self, message: str, code: str):
        super().__init__(message, code, status_code=400, action="RETRY_LOGIN")


class OAuthNotConfiguredError(OpenGovException):
    """Raised when OAuth credentials are not configured"""
    def __init__(self):
        super().__init__(
            message="Authentication service is not configured. Please contact support.",
            code="OAUTH_NOT_CONFIGURED",
            status_code=500,
            action="CONTACT_SUPPORT",
        )


class OAuthCodeExchangeError(OAuthError):
    """Raised when OAuth code exchange fails"""
    def __init__(self, details: Optional[str] = None):
        message = "Authentication failed. Please try again."
        if details:
            message = f"{message} ({details})"
        super().__init__(
            message=message,
            code="OAUTH_EXCHANGE_FAILED",
        )


class OAuthUserInfoError(OAuthError):
    """Raised when fetching user info from OAuth provider fails"""
    def __init__(self):
        super().__init__(
            message="Failed to retrieve your account information. Please try again.",
            code="OAUTH_USERINFO_FAILED",
        )


# API Exceptions
class APIError(OpenGovException):
    """Base class for API errors"""
    pass


class ValidationError(APIError):
    """Raised for request validation errors"""
    def __init__(self, message: str, fields: Optional[dict] = None):
        super().__init__(
            message=message,
            code="VALIDATION_ERROR",
            status_code=422,
        )
        self.fields = fields or {}
