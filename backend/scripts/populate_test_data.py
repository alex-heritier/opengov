#!/usr/bin/env python3
"""
Script to populate database with test articles for performance testing.
Usage: python -m scripts.populate_test_data
"""
import time
from datetime import datetime, timedelta, timezone

from app.database import SessionLocal, engine, Base
from app.models import FRArticle


def populate_test_data(num_articles: int = 150):
    """Populate database with test FRArticles."""
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()

    try:
        # Check if data already exists
        existing_count = db.query(FRArticle).count()
        if existing_count > 0:
            print(f"Database already has {existing_count} articles. Skipping population.")
            return

        articles = []
        base_date = datetime.now(timezone.utc)

        for i in range(num_articles):
            published_at = base_date - timedelta(minutes=i*5)
            fetched_at = base_date - timedelta(minutes=i*5 + 2)  # Fetched 2 mins before published

            article = FRArticle(
                # Federal Register raw data
                document_number=f"2025-{i+1:06d}",
                raw_data={
                    "title": f"Federal Register Update {i+1}: Important Government Action {i+1}",
                    "abstract": f"Test summary for article {i+1}",
                    "document_number": f"2025-{i+1:06d}",
                    "publication_date": published_at.isoformat(),
                },
                fetched_at=fetched_at,
                # Processed article data
                title=f"Federal Register Update {i+1}: Important Government Action {i+1}",
                summary=f"This is a test summary for article {i+1}. "
                        f"Learn about critical government decisions and policy updates "
                        f"that affect millions of Americans. Stay informed about the latest "
                        f"Federal Register announcements and regulatory changes.",
                source_url=f"https://www.federalregister.gov/documents/2025/11/{13+i//30:02d}/doc-{i+1:06d}",
                published_at=published_at,
            )
            articles.append(article)

        # Bulk insert
        db.bulk_save_objects(articles)
        db.commit()
        print(f"Successfully inserted {num_articles} test articles")

        # Verify
        count = db.query(FRArticle).count()
        print(f"Total articles in database: {count}")

        # Test query performance
        print("\nPerformance Test Results:")

        # Test 1: Get paginated feed (page 1, 20 items)
        start = time.time()
        _ = db.query(FRArticle).order_by(
            FRArticle.published_at.desc()
        ).limit(20).offset(0).all()
        elapsed = time.time() - start
        print(f"  Paginated query (20 items): {elapsed*1000:.2f}ms")

        # Test 2: Get paginated feed (page 5, 20 items)
        start = time.time()
        _ = db.query(FRArticle).order_by(
            FRArticle.published_at.desc()
        ).limit(20).offset(80).all()
        elapsed = time.time() - start
        print(f"  Pagination offset 80: {elapsed*1000:.2f}ms")

        # Test 3: Get single article by ID
        start = time.time()
        _ = db.query(FRArticle).filter(FRArticle.id == 50).first()
        elapsed = time.time() - start
        print(f"  Single article lookup: {elapsed*1000:.2f}ms")

        # Test 4: Count all articles
        start = time.time()
        count = db.query(FRArticle).count()
        elapsed = time.time() - start
        print(f"  Count all articles: {elapsed*1000:.2f}ms")

    finally:
        db.close()


if __name__ == "__main__":
    populate_test_data(150)
