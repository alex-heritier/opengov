#!/usr/bin/env python3
"""
Script to populate database with test articles for performance testing.
Usage: python -m scripts.populate_test_data
"""
import asyncio
import sys
from datetime import datetime, timedelta
from sqlalchemy.orm import Session

# Add backend to path
sys.path.insert(0, '/Users/alex/Documents/Project/opengov/backend')

from app.database import SessionLocal, engine, Base
from app.models.article import Article


def populate_test_data(num_articles: int = 150):
    """Populate database with test articles."""
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()
    
    try:
        # Check if data already exists
        existing_count = db.query(Article).count()
        if existing_count > 0:
            print(f"Database already has {existing_count} articles. Skipping population.")
            return
        
        articles = []
        base_date = datetime.utcnow()
        
        for i in range(num_articles):
            published_at = base_date - timedelta(minutes=i*5)
            article = Article(
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
        count = db.query(Article).count()
        print(f"Total articles in database: {count}")
        
        # Test query performance
        print("\nPerformance Test Results:")
        
        # Test 1: Get paginated feed (page 1, 20 items)
        import time
        start = time.time()
        results = db.query(Article).order_by(Article.published_at.desc()).limit(20).offset(0).all()
        elapsed = time.time() - start
        print(f"  Paginated query (20 items): {elapsed*1000:.2f}ms")
        
        # Test 2: Get paginated feed (page 5, 20 items)
        start = time.time()
        results = db.query(Article).order_by(Article.published_at.desc()).limit(20).offset(80).all()
        elapsed = time.time() - start
        print(f"  Pagination offset 80: {elapsed*1000:.2f}ms")
        
        # Test 3: Get single article by ID
        start = time.time()
        results = db.query(Article).filter(Article.id == 50).first()
        elapsed = time.time() - start
        print(f"  Single article lookup: {elapsed*1000:.2f}ms")
        
        # Test 4: Count all articles
        start = time.time()
        count = db.query(Article).count()
        elapsed = time.time() - start
        print(f"  Count all articles: {elapsed*1000:.2f}ms")
        
    finally:
        db.close()


if __name__ == "__main__":
    populate_test_data(150)
