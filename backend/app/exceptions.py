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
