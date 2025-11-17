import pytest
from datetime import datetime, timezone
from httpx import AsyncClient
from sqlalchemy.orm import Session
from app.models import FRArticle


@pytest.fixture
def setup_test_data(db_session: Session):
    """Setup test articles"""
    # Create unified FRArticle objects (combining raw and processed data)
    articles = [
        FRArticle(
            document_number="TEST-2024-001",
            raw_data={"test": "data", "title": "Test Article 1"},
            fetched_at=datetime.now(timezone.utc),
            title="Test Article 1",
            summary="This is a test summary",
            source_url="https://example.com/1",
            published_at=datetime.now(timezone.utc),
        ),
        FRArticle(
            document_number="TEST-2024-002",
            raw_data={"test": "data2", "title": "Test Article 2"},
            fetched_at=datetime.now(timezone.utc),
            title="Test Article 2",
            summary="Another test summary",
            source_url="https://example.com/2",
            published_at=datetime.now(timezone.utc),
        ),
    ]

    for article in articles:
        db_session.add(article)

    db_session.commit()

    yield

    # Cleanup is handled by conftest.py dropping all tables after each test


@pytest.mark.asyncio
async def test_get_feed(client: AsyncClient, setup_test_data):
    """Test getting feed"""
    response = await client.get("/api/feed")
    assert response.status_code == 200
    data = response.json()
    assert data["page"] == 1
    assert data["limit"] == 20
    assert len(data["articles"]) == 2


@pytest.mark.asyncio
async def test_get_feed_with_pagination(client: AsyncClient, setup_test_data):
    """Test feed pagination"""
    response = await client.get("/api/feed?page=1&limit=1")
    assert response.status_code == 200
    data = response.json()
    assert len(data["articles"]) == 1
    assert data["has_next"] is True


@pytest.mark.asyncio
async def test_get_article_not_found(client: AsyncClient):
    """Test getting non-existent article"""
    response = await client.get("/api/feed/999")
    assert response.status_code == 404
