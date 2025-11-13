"""
Tests for scraper run monitoring and tracking.
"""
import pytest
from datetime import datetime, timezone
from unittest.mock import Mock, patch, AsyncMock
from sqlalchemy.orm import Session

from app.models import ScraperRun
from app.schemas import ScraperRunResponse, ScraperRunListResponse
from app.workers.scraper import fetch_and_process


@pytest.mark.asyncio
async def test_scraper_run_creation():
    """Test that scraper creates a ScraperRun record"""
    # Mock the database and services
    with patch('app.workers.scraper.SessionLocal') as mock_session_local:
        mock_db = Mock(spec=Session)
        mock_session_local.return_value = mock_db

        # Mock the query to check for running scrapers (return None = no running scraper)
        mock_query = Mock()
        mock_query.filter.return_value.first.return_value = None
        mock_db.query.return_value = mock_query

        # Mock fetch_recent_documents to return empty list (no documents)
        with patch('app.workers.scraper.fetch_recent_documents', new_callable=AsyncMock) as mock_fetch:
            mock_fetch.return_value = []

            # Run scraper
            await fetch_and_process()

            # Verify ScraperRun was created
            assert mock_db.add.called
            assert mock_db.commit.called


def test_scraper_run_response_schema():
    """Test ScraperRunResponse schema"""
    now = datetime.now(timezone.utc)
    later = datetime.now(timezone.utc)
    
    run_data = {
        'id': 1,
        'started_at': now,
        'completed_at': later,
        'processed_count': 10,
        'skipped_count': 5,
        'error_count': 0,
        'success': True,
        'error_message': None
    }
    
    response = ScraperRunResponse(**run_data)
    
    assert response.id == 1
    assert response.processed_count == 10
    assert response.skipped_count == 5
    assert response.error_count == 0
    assert response.success is True
    assert response.duration_seconds is not None


def test_scraper_run_list_response_schema():
    """Test ScraperRunListResponse schema"""
    now = datetime.now(timezone.utc)
    
    run1 = ScraperRunResponse(
        id=1,
        started_at=now,
        completed_at=now,
        processed_count=10,
        skipped_count=5,
        error_count=0,
        success=True,
        error_message=None
    )
    
    response = ScraperRunListResponse(runs=[run1], total=1)
    
    assert len(response.runs) == 1
    assert response.total == 1
    assert response.runs[0].id == 1


def test_scraper_run_duration_calculation():
    """Test that duration_seconds is calculated correctly"""
    from datetime import timedelta

    now = datetime.now(timezone.utc)
    later = now + timedelta(seconds=45.5)
    
    response = ScraperRunResponse(
        id=1,
        started_at=now,
        completed_at=later,
        processed_count=10,
        skipped_count=5,
        error_count=0,
        success=True,
        error_message=None
    )
    
    assert response.duration_seconds == pytest.approx(45.5, abs=0.1)


def test_scraper_run_no_completion():
    """Test duration_seconds when job hasn't completed"""
    now = datetime.now(timezone.utc)
    
    response = ScraperRunResponse(
        id=1,
        started_at=now,
        completed_at=None,
        processed_count=0,
        skipped_count=0,
        error_count=0,
        success=False,
        error_message=None
    )
    
    assert response.duration_seconds is None
