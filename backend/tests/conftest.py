"""Test configuration and fixtures for pytest"""
import asyncio
import pytest
from typing import AsyncGenerator
from httpx import AsyncClient, ASGITransport
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker

# Import database first
from app.database import Base

# Import all models to register them with Base.metadata BEFORE creating engine
from app.models import Article, FederalRegister, Agency, User  # noqa: F401

# Import app and other dependencies
from app.main import app
from app.auth import get_async_session
from fastapi_users.password import PasswordHelper


# Test database URL - use in-memory SQLite for tests
# Using a shared in-memory database with a URI and shared cache
TEST_DATABASE_URL_ASYNC = "sqlite+aiosqlite:///file::memory:?cache=shared&uri=true"
TEST_DATABASE_URL_SYNC = "sqlite:///file::memory:?cache=shared&uri=true"

# Create test engines with StaticPool to share the same connection
from sqlalchemy.pool import StaticPool
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, Session

# Async engine for fastapi-users
test_async_engine = create_async_engine(
    TEST_DATABASE_URL_ASYNC,
    connect_args={"check_same_thread": False, "uri": True},
    poolclass=StaticPool,
    echo=False,
)

# Sync engine for feed routes (since they use synchronous DB)
test_sync_engine = create_engine(
    TEST_DATABASE_URL_SYNC,
    connect_args={"check_same_thread": False, "uri": True},
    poolclass=StaticPool,
    echo=False,
)

# Create session makers
TestAsyncSessionLocal = async_sessionmaker(
    test_async_engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

TestSyncSessionLocal = sessionmaker(
    autocommit=False,
    autoflush=False,
    bind=test_sync_engine,
)


@pytest.fixture(scope="session")
def event_loop():
    """Create an instance of the default event loop for the test session."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
def db_session() -> Session:
    """Create a fresh database session for each test (synchronous for feed tests)"""
    # Create all tables in sync engine
    Base.metadata.create_all(test_sync_engine)

    # Create session
    session = TestSyncSessionLocal()
    try:
        yield session
    finally:
        session.close()

    # Drop all tables after test
    Base.metadata.drop_all(test_sync_engine)


@pytest.fixture
async def client(db_session: Session) -> AsyncGenerator[AsyncClient, None]:
    """Create a test client with database session override"""
    from app.routers.common import get_db

    def override_get_db():
        """Override database dependency for testing"""
        try:
            yield db_session
        finally:
            pass

    async def override_get_async_session() -> AsyncGenerator[AsyncSession, None]:
        """Override async session for fastapi-users"""
        async with TestAsyncSessionLocal() as session:
            yield session

    app.dependency_overrides[get_db] = override_get_db
    app.dependency_overrides[get_async_session] = override_get_async_session

    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test"
    ) as ac:
        yield ac

    app.dependency_overrides.clear()


@pytest.fixture
async def password_helper() -> PasswordHelper:
    """Create a password helper for hashing passwords"""
    return PasswordHelper()


@pytest.fixture
async def test_user(db_session: AsyncSession, password_helper: PasswordHelper) -> User:
    """Create a test user in the database"""
    user = User(
        email="test@example.com",
        hashed_password=password_helper.hash("testpassword123"),
        is_active=True,
        is_verified=True,
        is_superuser=False,
        name="Test User",
    )
    db_session.add(user)
    await db_session.commit()
    await db_session.refresh(user)
    return user


@pytest.fixture
async def test_superuser(db_session: AsyncSession, password_helper: PasswordHelper) -> User:
    """Create a test superuser in the database"""
    user = User(
        email="admin@example.com",
        hashed_password=password_helper.hash("adminpassword123"),
        is_active=True,
        is_verified=True,
        is_superuser=True,
        name="Admin User",
    )
    db_session.add(user)
    await db_session.commit()
    await db_session.refresh(user)
    return user


@pytest.fixture
async def authenticated_client(
    client: AsyncClient,
    test_user: User
) -> AsyncGenerator[AsyncClient, None]:
    """Create an authenticated test client with valid session cookie"""
    # Login to get authentication cookie
    response = await client.post(
        "/api/auth/login",
        data={
            "username": test_user.email,
            "password": "testpassword123",
        },
    )
    assert response.status_code == 204, f"Login failed: {response.text}"

    # The cookie is automatically stored in the client
    yield client
