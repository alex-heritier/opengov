import pytest
from datetime import datetime
from fastapi.testclient import TestClient
from app.main import app
from app.database import SessionLocal, Base, engine
from app.models import Article

# Create test database
Base.metadata.create_all(bind=engine)

client = TestClient(app)


@pytest.fixture
def setup_test_data(autouse=False):
    """Setup test articles"""
    db = SessionLocal()
    
    # Clear existing data
    db.query(Article).delete()
    
    # Add test articles
    articles = [
        Article(
            title="Test Article 1",
            summary="This is a test summary",
            source_url="https://example.com/1",
            published_at=datetime.utcnow(),
        ),
        Article(
            title="Test Article 2",
            summary="Another test summary",
            source_url="https://example.com/2",
            published_at=datetime.utcnow(),
        ),
    ]
    
    for article in articles:
        db.add(article)
    
    db.commit()
    
    yield
    
    # Cleanup
    db.query(Article).delete()
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
