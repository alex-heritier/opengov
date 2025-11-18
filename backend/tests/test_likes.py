"""Test cases for like service and API endpoints"""
import pytest
from datetime import datetime, timezone
from httpx import AsyncClient
from sqlalchemy.orm import Session
from app.models import Like, FRArticle, User
from app.services import like as like_service


@pytest.fixture
def setup_like_test_data(db_session: Session):
    """Setup test articles and users for like tests"""
    # Create test articles
    articles = [
        FRArticle(
            document_number="LIKE-TEST-001",
            raw_data={"test": "data1"},
            fetched_at=datetime.now(timezone.utc),
            title="Like Test Article 1",
            summary="Test article for liking",
            source_url="https://example.com/like1",
            published_at=datetime.now(timezone.utc),
        ),
        FRArticle(
            document_number="LIKE-TEST-002",
            raw_data={"test": "data2"},
            fetched_at=datetime.now(timezone.utc),
            title="Like Test Article 2",
            summary="Another article to like",
            source_url="https://example.com/like2",
            published_at=datetime.now(timezone.utc),
        ),
        FRArticle(
            document_number="LIKE-TEST-003",
            raw_data={"test": "data3"},
            fetched_at=datetime.now(timezone.utc),
            title="Like Test Article 3",
            summary="Article with multiple likes",
            source_url="https://example.com/like3",
            published_at=datetime.now(timezone.utc),
        ),
    ]

    for article in articles:
        db_session.add(article)

    db_session.commit()
    yield
    # Cleanup handled by conftest.py


class TestLikeService:
    """Test like service functions"""

    @pytest.mark.asyncio
    async def test_toggle_like_create_positive(self, db_session: Session, setup_like_test_data):
        """Test creating a new positive like"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        like = like_service.toggle_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id,
            is_positive=True
        )
        
        assert like is not None
        assert like.user_id == 1
        assert like.frarticle_id == article.id
        assert like.is_positive is True

    @pytest.mark.asyncio
    async def test_toggle_like_create_negative(self, db_session: Session, setup_like_test_data):
        """Test creating a new dislike"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        like = like_service.toggle_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id,
            is_positive=False
        )
        
        assert like is not None
        assert like.is_positive is False

    @pytest.mark.asyncio
    async def test_toggle_like_remove_same_vote(self, db_session: Session, setup_like_test_data):
        """Test removing like by voting the same way again"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        # Create like
        like1 = like_service.toggle_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id,
            is_positive=True
        )
        assert like1 is not None
        
        # Toggle same vote - should remove
        like2 = like_service.toggle_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id,
            is_positive=True
        )
        
        assert like2 is None
        
        # Verify it's actually removed
        status = like_service.get_like_status(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        assert status is None

    @pytest.mark.asyncio
    async def test_toggle_like_change_vote(self, db_session: Session, setup_like_test_data):
        """Test changing vote from like to dislike"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        # Create like
        like1 = like_service.toggle_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id,
            is_positive=True
        )
        
        # Change to dislike
        like2 = like_service.toggle_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id,
            is_positive=False
        )
        
        assert like2 is not None
        assert like2.id == like1.id  # Same like object
        assert like2.is_positive is False

    @pytest.mark.asyncio
    async def test_toggle_like_nonexistent_article(self, db_session: Session):
        """Test liking a non-existent article"""
        with pytest.raises(ValueError, match="Article with ID 999 not found"):
            like_service.toggle_like(
                db=db_session,
                user_id=1,
                frarticle_id=999,
                is_positive=True
            )

    @pytest.mark.asyncio
    async def test_get_like_status_positive(self, db_session: Session, setup_like_test_data):
        """Test getting like status when liked"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        like_service.toggle_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id,
            is_positive=True
        )
        
        status = like_service.get_like_status(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert status is True

    @pytest.mark.asyncio
    async def test_get_like_status_negative(self, db_session: Session, setup_like_test_data):
        """Test getting like status when disliked"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        like_service.toggle_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id,
            is_positive=False
        )
        
        status = like_service.get_like_status(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert status is False

    @pytest.mark.asyncio
    async def test_get_like_status_none(self, db_session: Session, setup_like_test_data):
        """Test getting like status when not voted"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        status = like_service.get_like_status(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert status is None

    @pytest.mark.asyncio
    async def test_remove_like_existing(self, db_session: Session, setup_like_test_data):
        """Test removing an existing like"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        like_service.toggle_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id,
            is_positive=True
        )
        
        success = like_service.remove_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert success is True
        
        # Verify it's removed
        status = like_service.get_like_status(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        assert status is None

    @pytest.mark.asyncio
    async def test_remove_like_nonexistent(self, db_session: Session, setup_like_test_data):
        """Test removing a non-existent like"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        success = like_service.remove_like(
            db=db_session,
            user_id=1,
            frarticle_id=article.id
        )
        
        assert success is False

    @pytest.mark.asyncio
    async def test_get_article_like_counts_no_votes(self, db_session: Session, setup_like_test_data):
        """Test getting like counts for article with no votes"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        counts = like_service.get_article_like_counts(
            db=db_session,
            frarticle_id=article.id
        )
        
        assert counts["likes"] == 0
        assert counts["dislikes"] == 0

    @pytest.mark.asyncio
    async def test_get_article_like_counts_with_votes(self, db_session: Session, setup_like_test_data):
        """Test getting like counts with multiple votes"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-003").first()
        
        # Add multiple likes and dislikes from different users
        like_service.toggle_like(db=db_session, user_id=1, frarticle_id=article.id, is_positive=True)
        like_service.toggle_like(db=db_session, user_id=2, frarticle_id=article.id, is_positive=True)
        like_service.toggle_like(db=db_session, user_id=3, frarticle_id=article.id, is_positive=False)
        like_service.toggle_like(db=db_session, user_id=4, frarticle_id=article.id, is_positive=False)
        like_service.toggle_like(db=db_session, user_id=5, frarticle_id=article.id, is_positive=False)
        
        counts = like_service.get_article_like_counts(
            db=db_session,
            frarticle_id=article.id
        )
        
        assert counts["likes"] == 2
        assert counts["dislikes"] == 3


class TestLikeAPI:
    """Test like API endpoints"""

    @pytest.mark.asyncio
    async def test_toggle_like_endpoint_positive(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_like_test_data):
        """Test POST /api/likes/toggle endpoint with positive vote"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        response = await authenticated_client_sync.post(
            "/api/likes/toggle",
            json={"frarticle_id": article.id, "is_positive": True}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data is not None
        assert data["is_positive"] is True

    @pytest.mark.asyncio
    async def test_toggle_like_endpoint_negative(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_like_test_data):
        """Test POST /api/likes/toggle endpoint with negative vote"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        response = await authenticated_client_sync.post(
            "/api/likes/toggle",
            json={"frarticle_id": article.id, "is_positive": False}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["is_positive"] is False

    @pytest.mark.asyncio
    async def test_toggle_like_endpoint_remove(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_like_test_data):
        """Test POST /api/likes/toggle removing a like"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        # Create like
        await authenticated_client_sync.post(
            "/api/likes/toggle",
            json={"frarticle_id": article.id, "is_positive": True}
        )
        
        # Remove like by voting same way
        response = await authenticated_client_sync.post(
            "/api/likes/toggle",
            json={"frarticle_id": article.id, "is_positive": True}
        )
        
        assert response.status_code == 200
        assert response.json() is None

    @pytest.mark.asyncio
    async def test_toggle_like_endpoint_nonexistent_article(self, authenticated_client_sync: AsyncClient):
        """Test toggle endpoint with non-existent article"""
        response = await authenticated_client_sync.post(
            "/api/likes/toggle",
            json={"frarticle_id": 999, "is_positive": True}
        )
        
        assert response.status_code == 404

    @pytest.mark.asyncio
    async def test_remove_like_endpoint(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_like_test_data):
        """Test DELETE /api/likes/{id} endpoint"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        user = db_session.query(User).filter_by(email="test@example.com").first()
        
        like_service.toggle_like(
            db=db_session,
            user_id=user.id,
            frarticle_id=article.id,
            is_positive=True
        )
        
        response = await authenticated_client_sync.delete(f"/api/likes/{article.id}")
        
        assert response.status_code == 200
        assert response.json()["success"] is True

    @pytest.mark.asyncio
    async def test_remove_like_endpoint_nonexistent(self, authenticated_client_sync: AsyncClient):
        """Test DELETE endpoint with non-existent like"""
        response = await authenticated_client_sync.delete("/api/likes/999")
        
        assert response.status_code == 404

    @pytest.mark.asyncio
    async def test_get_like_status_endpoint_positive(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_like_test_data):
        """Test GET /api/likes/status/{id} endpoint when liked"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        user = db_session.query(User).filter_by(email="test@example.com").first()
        
        like_service.toggle_like(
            db=db_session,
            user_id=user.id,
            frarticle_id=article.id,
            is_positive=True
        )
        
        response = await authenticated_client_sync.get(f"/api/likes/status/{article.id}")
        
        assert response.status_code == 200
        data = response.json()
        assert data["is_positive"] is True

    @pytest.mark.asyncio
    async def test_get_like_status_endpoint_negative(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_like_test_data):
        """Test GET /api/likes/status/{id} endpoint when disliked"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        user = db_session.query(User).filter_by(email="test@example.com").first()
        
        like_service.toggle_like(
            db=db_session,
            user_id=user.id,
            frarticle_id=article.id,
            is_positive=False
        )
        
        response = await authenticated_client_sync.get(f"/api/likes/status/{article.id}")
        
        assert response.status_code == 200
        data = response.json()
        assert data["is_positive"] is False

    @pytest.mark.asyncio
    async def test_get_like_status_endpoint_none(self, authenticated_client_sync: AsyncClient, db_session: Session, setup_like_test_data):
        """Test GET /api/likes/status/{id} endpoint when not voted"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        response = await authenticated_client_sync.get(f"/api/likes/status/{article.id}")
        
        assert response.status_code == 200
        data = response.json()
        assert data["is_positive"] is None

    @pytest.mark.asyncio
    async def test_get_like_counts_endpoint(self, client: AsyncClient, db_session: Session, setup_like_test_data):
        """Test GET /api/likes/counts/{id} endpoint (no auth required)"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-003").first()
        
        # Add votes
        like_service.toggle_like(db=db_session, user_id=1, frarticle_id=article.id, is_positive=True)
        like_service.toggle_like(db=db_session, user_id=2, frarticle_id=article.id, is_positive=True)
        like_service.toggle_like(db=db_session, user_id=3, frarticle_id=article.id, is_positive=False)
        
        response = await client.get(f"/api/likes/counts/{article.id}")
        
        assert response.status_code == 200
        data = response.json()
        assert data["likes"] == 2
        assert data["dislikes"] == 1

    @pytest.mark.asyncio
    async def test_like_endpoint_requires_auth(self, client: AsyncClient, db_session: Session, setup_like_test_data):
        """Test that like endpoints require authentication"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        response = await client.post(
            "/api/likes/toggle",
            json={"frarticle_id": article.id, "is_positive": True}
        )
        
        assert response.status_code == 401

    @pytest.mark.asyncio
    async def test_get_like_counts_no_auth_required(self, client: AsyncClient, db_session: Session, setup_like_test_data):
        """Test that getting like counts doesn't require authentication"""
        article = db_session.query(FRArticle).filter_by(document_number="LIKE-TEST-001").first()
        
        response = await client.get(f"/api/likes/counts/{article.id}")
        
        assert response.status_code == 200
