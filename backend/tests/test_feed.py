import pytest
from datetime import datetime, timezone
from fastapi.testclient import TestClient
from app.main import app
from app.database import SessionLocal, Base, engine
from app.models import Article
from app.models.federal_register import FederalRegister

# Create test database
Base.metadata.create_all(bind=engine)

client = TestClient(app)


@pytest.fixture
def setup_test_data(autouse=False):
    """Setup test articles"""
    db = SessionLocal()

    # Clear existing data
    db.query(Article).delete()
    db.query(FederalRegister).delete()

    # Create federal register entries first
    fed_entry1 = FederalRegister(
        document_number="TEST-2024-001",
        raw_data={"test": "data"},
        fetched_at=datetime.now(timezone.utc),
        processed=True
    )
    fed_entry2 = FederalRegister(
        document_number="TEST-2024-002",
        raw_data={"test": "data2"},
        fetched_at=datetime.now(timezone.utc),
        processed=True
    )
    db.add_all([fed_entry1, fed_entry2])
    db.flush()

    # Then create articles with foreign key references
    articles = [
        Article(
            federal_register_id=fed_entry1.id,
            title="Test Article 1",
            summary="This is a test summary",
            source_url="https://example.com/1",
            published_at=datetime.now(timezone.utc),
        ),
        Article(
            federal_register_id=fed_entry2.id,
            title="Test Article 2",
            summary="Another test summary",
            source_url="https://example.com/2",
            published_at=datetime.now(timezone.utc),
        ),
    ]

    for article in articles:
        db.add(article)

    db.commit()

    yield

    # Cleanup
    db.query(Article).delete()
    db.query(FederalRegister).delete()
    db.commit()
    db.close()


def test_get_feed(setup_test_data):
    """Test getting feed"""
    response = client.get("/api/feed")
    assert response.status_code == 200
    data = response.json()
    assert data["page"] == 1
    assert data["limit"] == 20
    assert len(data["articles"]) == 2


def test_get_feed_with_pagination(setup_test_data):
    """Test feed pagination"""
    response = client.get("/api/feed?page=1&limit=1")
    assert response.status_code == 200
    data = response.json()
    assert len(data["articles"]) == 1
    assert data["has_next"] is True


def test_get_article_not_found():
    """Test getting non-existent article"""
    response = client.get("/api/feed/999")
    assert response.status_code == 404
