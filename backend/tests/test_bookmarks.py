"""Test cases for bookmark service and API endpoints"""
import pytest
from datetime import datetime, timezone
from httpx import AsyncClient
from sqlalchemy.orm import Session
from app.models import Bookmark, FRArticle, User
from app.services import bookmark as bookmark_service


@pytest.fixture
def setup_bookmark_test_data(db_session: Session):
    """Setup test articles and users for bookmark tests"""
    # Create test articles
    articles = [
        FRArticle(
            document_number="BOOKMARK-TEST-001",
            raw_data={"test": "data1"},
            fetched_at=datetime.now(timezone.utc),
            title="Bookmark Test Article 1",
            summary="Test summary for bookmarking",
            source_url="https://example.com/bookmark1",
            published_at=datetime.now(timezone.utc),
        ),
        FRArticle(
            document_number="BOOKMARK-TEST-002",
            raw_data={"test": "data2"},
            fetched_at=datetime.now(timezone.utc),
            title="Bookmark Test Article 2",
            summary="Another article to bookmark",
            source_url="https://example.com/bookmark2",
            published_at=datetime.now(timezone.utc),
        ),
    ]

    for article in articles:
        db_session.add(article)

    db_session.commit()
    yield
    # Cleanup handled by conftest.py


class TestBookmarkService:
    """Test bookmark service functions"""

    @pytest.mark.asyncio
    async def test_toggle_bookmark_create_new(self, db_session: Session, setup_bookmark_test_data):
        """Test creating a new bookmark"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        
        bookmark = bookmark_service.toggle_bookmark(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert bookmark.user_id == 1
        assert bookmark.frarticle_id == article.id
        assert bookmark.is_bookmarked is True
        assert bookmark.created_at is not None

    @pytest.mark.asyncio
    async def test_toggle_bookmark_toggle_existing(self, db_session: Session, setup_bookmark_test_data):
        """Test toggling an existing bookmark"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        
        # Create initial bookmark
        bookmark1 = bookmark_service.toggle_bookmark(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        assert bookmark1.is_bookmarked is True
        
        # Toggle it
        bookmark2 = bookmark_service.toggle_bookmark(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert bookmark2.is_bookmarked is False
        assert bookmark2.id == bookmark1.id

    @pytest.mark.asyncio
    async def test_toggle_bookmark_nonexistent_article(self, db_session: Session):
        """Test bookmarking a non-existent article"""
        with pytest.raises(ValueError, match="Article with ID 999 not found"):
            bookmark_service.toggle_bookmark(
                db=db_session,
                user_id=1,
                frarticle_id=999
            )

    @pytest.mark.asyncio
    async def test_get_user_bookmarks(self, db_session: Session, setup_bookmark_test_data):
        """Test retrieving user bookmarks"""
        articles = db_session.query(FRArticle).all()
        
        # Create bookmarks
        bookmark_service.toggle_bookmark(db=db_session, user_id=1, frarticle_id=articles[0].id)
        bookmark_service.toggle_bookmark(db=db_session, user_id=1, frarticle_id=articles[1].id)
        
        bookmarks = bookmark_service.get_user_bookmarks(db=db_session, user_id=1)
        
        assert len(bookmarks) == 2
        assert all(b.is_bookmarked for b in bookmarks)

    @pytest.mark.asyncio
    async def test_get_user_bookmarks_exclude_unbookmarked(self, db_session: Session, setup_bookmark_test_data):
        """Test excluding unbookmarked items from results"""
        articles = db_session.query(FRArticle).all()
        
        # Create and toggle bookmark
        b = bookmark_service.toggle_bookmark(db=db_session, user_id=1, frarticle_id=articles[0].id)
        bookmark_service.toggle_bookmark(db=db_session, user_id=1, frarticle_id=articles[0].id)
        
        # With include_unbookmarked=False (default)
        bookmarks = bookmark_service.get_user_bookmarks(
            db=db_session,
            user_id=1,
            include_unbookmarked=False
        )
        assert len(bookmarks) == 0
        
        # With include_unbookmarked=True
        bookmarks = bookmark_service.get_user_bookmarks(
            db=db_session,
            user_id=1,
            include_unbookmarked=True
        )
        assert len(bookmarks) == 1
        assert bookmarks[0].is_bookmarked is False

    @pytest.mark.asyncio
    async def test_get_bookmark_status_bookmarked(self, db_session: Session, setup_bookmark_test_data):
        """Test checking bookmark status when bookmarked"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        
        bookmark_service.toggle_bookmark(db=db_session, user_id=1, frarticle_id=article.id)
        
        is_bookmarked = bookmark_service.get_bookmark_status(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert is_bookmarked is True

    @pytest.mark.asyncio
    async def test_get_bookmark_status_not_bookmarked(self, db_session: Session, setup_bookmark_test_data):
        """Test checking bookmark status when not bookmarked"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        
        is_bookmarked = bookmark_service.get_bookmark_status(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert is_bookmarked is False

    @pytest.mark.asyncio
    async def test_get_bookmark_status_after_toggle_off(self, db_session: Session, setup_bookmark_test_data):
        """Test bookmark status after toggling off"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        
        bookmark_service.toggle_bookmark(db=db_session, user_id=1, frarticle_id=article.id)
        bookmark_service.toggle_bookmark(db=db_session, user_id=1, frarticle_id=article.id)
        
        is_bookmarked = bookmark_service.get_bookmark_status(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert is_bookmarked is False

    @pytest.mark.asyncio
    async def test_get_bookmarked_articles(self, db_session: Session, setup_bookmark_test_data):
        """Test retrieving bookmarked articles"""
        articles = db_session.query(FRArticle).all()
        
        bookmark_service.toggle_bookmark(db=db_session, user_id=1, frarticle_id=articles[0].id)
        bookmark_service.toggle_bookmark(db=db_session, user_id=1, frarticle_id=articles[1].id)
        
        bookmarked = bookmark_service.get_bookmarked_articles(db=db_session, user_id=1)
        
        assert len(bookmarked) == 2
        assert any(a.document_number == "BOOKMARK-TEST-001" for a in bookmarked)
        assert any(a.document_number == "BOOKMARK-TEST-002" for a in bookmarked)

    @pytest.mark.asyncio
    async def test_get_bookmarked_articles_empty(self, db_session: Session, setup_bookmark_test_data):
        """Test retrieving bookmarked articles when none exist"""
        bookmarked = bookmark_service.get_bookmarked_articles(db=db_session, user_id=999)
        
        assert len(bookmarked) == 0

    @pytest.mark.asyncio
    async def test_remove_bookmark_existing(self, db_session: Session, setup_bookmark_test_data):
        """Test removing an existing bookmark"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        
        bookmark_service.toggle_bookmark(db=db_session, user_id=1, frarticle_id=article.id)
        
        success = bookmark_service.remove_bookmark(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert success is True
        
        # Verify it's marked as unbookmarked
        is_bookmarked = bookmark_service.get_bookmark_status(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        assert is_bookmarked is False

    @pytest.mark.asyncio
    async def test_remove_bookmark_nonexistent(self, db_session: Session, setup_bookmark_test_data):
        """Test removing a non-existent bookmark"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        
        success = bookmark_service.remove_bookmark(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert success is False


class TestBookmarkAPI:
    """Test bookmark API endpoints"""

    @pytest.mark.asyncio
    async def test_toggle_bookmark_endpoint(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_bookmark_test_data):
        """Test POST /api/bookmarks/toggle endpoint"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        
        response = await authenticated_client_sync.post(
            "/api/bookmarks/toggle",
            json={"frarticle_id": article.id}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["is_bookmarked"] is True

    @pytest.mark.asyncio
    async def test_toggle_bookmark_endpoint_nonexistent_article(self, authenticated_client_sync: AsyncClient):
        """Test toggle endpoint with non-existent article"""
        response = await authenticated_client_sync.post(
            "/api/bookmarks/toggle",
            json={"frarticle_id": 999}
        )
        
        assert response.status_code == 404

    @pytest.mark.asyncio
    async def test_get_bookmarks_endpoint(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_bookmark_test_data):
        """Test GET /api/bookmarks endpoint"""
        articles = db_session.query(FRArticle).all()
        user = db_session.query(User).filter_by(email="test@example.com").first()
        
        # Create bookmarks
        for article in articles:
            bookmark_service.toggle_bookmark(
                db=db_session,
                user_id=user.id,
                frarticle_id=article.id
            )
        
        response = await authenticated_client_sync.get("/api/bookmarks")
        
        assert response.status_code == 200
        data = response.json()
        assert len(data) == 2
        assert all("id" in item and "title" in item for item in data)

    @pytest.mark.asyncio
    async def test_get_bookmarks_endpoint_empty(self, authenticated_client_sync: AsyncClient):
        """Test GET /api/bookmarks when no bookmarks exist"""
        response = await authenticated_client_sync.get("/api/bookmarks")
        
        assert response.status_code == 200
        data = response.json()
        assert len(data) == 0

    @pytest.mark.asyncio
    async def test_remove_bookmark_endpoint(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_bookmark_test_data):
        """Test DELETE /api/bookmarks/{id} endpoint"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        user = db_session.query(User).filter_by(email="test@example.com").first()
        
        bookmark_service.toggle_bookmark(
            db=db_session,
            user_id=user.id,
            frarticle_id=article.id
        )
        
        response = await authenticated_client_sync.delete(f"/api/bookmarks/{article.id}")
        
        assert response.status_code == 200
        assert response.json()["success"] is True

    @pytest.mark.asyncio
    async def test_remove_bookmark_endpoint_nonexistent(self, authenticated_client_sync: AsyncClient):
        """Test DELETE endpoint with non-existent bookmark"""
        response = await authenticated_client_sync.delete("/api/bookmarks/999")
        
        assert response.status_code == 404

    @pytest.mark.asyncio
    async def test_get_bookmark_status_endpoint(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_bookmark_test_data):
        """Test GET /api/bookmarks/status/{id} endpoint"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        user = db_session.query(User).filter_by(email="test@example.com").first()
        
        bookmark_service.toggle_bookmark(
            db=db_session,
            user_id=user.id,
            frarticle_id=article.id
        )
        
        response = await authenticated_client_sync.get(f"/api/bookmarks/status/{article.id}")
        
        assert response.status_code == 200
        data = response.json()
        assert data["is_bookmarked"] is True

    @pytest.mark.asyncio
    async def test_get_bookmark_status_endpoint_not_bookmarked(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_bookmark_test_data):
        """Test status endpoint when not bookmarked"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        
        response = await authenticated_client_sync.get(f"/api/bookmarks/status/{article.id}")
        
        assert response.status_code == 200
        data = response.json()
        assert data["is_bookmarked"] is False

    @pytest.mark.asyncio
    async def test_bookmark_endpoint_requires_auth(self, client: AsyncClient, db_session: Session, setup_bookmark_test_data):
        """Test that bookmark endpoints require authentication"""
        article = db_session.query(FRArticle).filter_by(document_number="BOOKMARK-TEST-001").first()
        
        response = await client.post(
            "/api/bookmarks/toggle",
            json={"frarticle_id": article.id}
        )
        
        assert response.status_code == 401
