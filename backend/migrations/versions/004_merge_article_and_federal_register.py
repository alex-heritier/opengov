"""Merge Article and FederalRegister into FRArticle

Revision ID: 004_merge_article_federal_register
Revises: 003_add_users_table
Create Date: 2025-11-17 00:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

# revision identifiers, used by Alembic.
revision: str = '004_merge_article_federal_register'
down_revision: Union[str, None] = '003_add_users_table'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """
    Merge articles and federal_register_entries tables into frarticles.

    Strategy:
    1. Create new frarticles table with combined schema
    2. Migrate data from both tables
    3. Drop old tables
    """

    # Create new frarticles table
    op.create_table(
        'frarticles',
        sa.Column('id', sa.Integer(), nullable=False),

        # Federal Register raw data fields
        sa.Column('document_number', sa.String(length=50), nullable=False),
        sa.Column('raw_data', sa.JSON(), nullable=False),
        sa.Column('fetched_at', sa.DateTime(), nullable=False),

        # Processed article fields
        sa.Column('title', sa.String(length=500), nullable=False),
        sa.Column('summary', sa.Text(), nullable=False),
        sa.Column('source_url', sa.String(length=500), nullable=False),
        sa.Column('published_at', sa.DateTime(), nullable=False),

        # Metadata
        sa.Column('created_at', sa.DateTime(), nullable=False),
        sa.Column('updated_at', sa.DateTime(), nullable=False),

        sa.PrimaryKeyConstraint('id')
    )

    # Create indexes
    op.create_index('ix_frarticles_id', 'frarticles', ['id'], unique=False)
    op.create_index(
        'ix_frarticles_document_number',
        'frarticles',
        ['document_number'],
        unique=True
    )
    op.create_index(
        'ix_frarticles_source_url',
        'frarticles',
        ['source_url'],
        unique=True
    )
    op.create_index(
        'ix_frarticles_published_at',
        'frarticles',
        ['published_at'],
        unique=False
    )
    op.create_index(
        'ix_frarticles_fetched_at',
        'frarticles',
        ['fetched_at'],
        unique=False
    )
    op.create_index(
        'idx_frarticles_published_at_desc',
        'frarticles',
        ['published_at'],
        unique=False
    )

    # Migrate data from old tables
    # This SQL joins articles with federal_register_entries and inserts into frarticles
    op.execute("""
        INSERT INTO frarticles (
            id,
            document_number,
            raw_data,
            fetched_at,
            title,
            summary,
            source_url,
            published_at,
            created_at,
            updated_at
        )
        SELECT
            a.id,
            fr.document_number,
            fr.raw_data,
            fr.fetched_at,
            a.title,
            a.summary,
            a.source_url,
            a.published_at,
            a.created_at,
            a.updated_at
        FROM articles a
        INNER JOIN federal_register_entries fr ON a.federal_register_id = fr.id
    """)

    # Drop old tables (indexes will be dropped automatically)
    op.drop_table('articles')
    op.drop_table('federal_register_entries')


def downgrade() -> None:
    """
    Reverse the merge: split frarticles back into articles and federal_register_entries.

    Note: This downgrade will work but may lose orphaned federal_register_entries
    that don't have corresponding articles.
    """

    # Recreate federal_register_entries table
    op.create_table(
        'federal_register_entries',
        sa.Column('id', sa.Integer(), nullable=False),
        sa.Column('document_number', sa.String(length=50), nullable=False),
        sa.Column('raw_data', sa.JSON(), nullable=False),
        sa.Column('fetched_at', sa.DateTime(), nullable=False),
        sa.Column('processed', sa.Boolean(), nullable=False),
        sa.PrimaryKeyConstraint('id')
    )

    # Create indexes for federal_register_entries
    op.create_index(
        'ix_federal_register_entries_id',
        'federal_register_entries',
        ['id'],
        unique=False
    )
    op.create_index(
        'ix_federal_register_entries_document_number',
        'federal_register_entries',
        ['document_number'],
        unique=True
    )
    op.create_index(
        'ix_federal_register_entries_processed',
        'federal_register_entries',
        ['processed'],
        unique=False
    )
    op.create_index(
        'idx_processed_fetched',
        'federal_register_entries',
        ['processed', 'fetched_at'],
        unique=False
    )

    # Recreate articles table
    op.create_table(
        'articles',
        sa.Column('id', sa.Integer(), nullable=False),
        sa.Column('federal_register_id', sa.Integer(), nullable=True),
        sa.Column('title', sa.String(length=500), nullable=False),
        sa.Column('summary', sa.Text(), nullable=False),
        sa.Column('source_url', sa.String(length=500), nullable=False),
        sa.Column('published_at', sa.DateTime(), nullable=False),
        sa.Column('created_at', sa.DateTime(), nullable=False),
        sa.Column('updated_at', sa.DateTime(), nullable=False),
        sa.ForeignKeyConstraint(['federal_register_id'], ['federal_register_entries.id'], ),
        sa.PrimaryKeyConstraint('id')
    )

    # Create indexes for articles
    op.create_index('ix_articles_id', 'articles', ['id'], unique=False)
    op.create_index(
        'ix_articles_federal_register_id',
        'articles',
        ['federal_register_id'],
        unique=False
    )
    op.create_index(
        'ix_articles_source_url',
        'articles',
        ['source_url'],
        unique=True
    )
    op.create_index(
        'ix_articles_published_at',
        'articles',
        ['published_at'],
        unique=False
    )
    op.create_index(
        'idx_published_at_desc',
        'articles',
        ['published_at'],
        unique=False
    )

    # Migrate data back from frarticles
    # First insert into federal_register_entries
    op.execute("""
        INSERT INTO federal_register_entries (
            id,
            document_number,
            raw_data,
            fetched_at,
            processed
        )
        SELECT
            id,
            document_number,
            raw_data,
            fetched_at,
            1 as processed
        FROM frarticles
    """)

    # Then insert into articles
    op.execute("""
        INSERT INTO articles (
            id,
            federal_register_id,
            title,
            summary,
            source_url,
            published_at,
            created_at,
            updated_at
        )
        SELECT
            id,
            id as federal_register_id,
            title,
            summary,
            source_url,
            published_at,
            created_at,
            updated_at
        FROM frarticles
    """)

    # Drop frarticles table
    op.drop_table('frarticles')
