"""Initial schema with federal_register_id

Revision ID: 001_initial_schema
Revises: 
Create Date: 2025-11-13 17:02:20.509411

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import sqlite

# revision identifiers, used by Alembic.
revision: str = '001_initial_schema'
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create federal_register_entries table
    op.create_table('federal_register_entries',
    sa.Column('id', sa.Integer(), nullable=False),
    sa.Column('document_number', sa.String(length=50), nullable=False),
    sa.Column('raw_data', sqlite.JSON(), nullable=False),
    sa.Column('fetched_at', sa.DateTime(), nullable=False),
    sa.Column('processed', sa.Boolean(), nullable=False),
    sa.PrimaryKeyConstraint('id'),
    sa.UniqueConstraint('document_number')
    )
    op.create_index('idx_processed_fetched', 'federal_register_entries', ['processed', 'fetched_at'], unique=False)
    op.create_index('ix_federal_register_entries_document_number', 'federal_register_entries', ['document_number'], unique=False)
    op.create_index('ix_federal_register_entries_processed', 'federal_register_entries', ['processed'], unique=False)

    # Create articles table with federal_register_id
    op.create_table('articles',
    sa.Column('id', sa.Integer(), nullable=False),
    sa.Column('federal_register_id', sa.Integer(), nullable=True),
    sa.Column('title', sa.String(length=500), nullable=False),
    sa.Column('summary', sa.Text(), nullable=False),
    sa.Column('source_url', sa.String(length=500), nullable=False),
    sa.Column('published_at', sa.DateTime(), nullable=False),
    sa.Column('created_at', sa.DateTime(), nullable=False),
    sa.Column('updated_at', sa.DateTime(), nullable=False),
    sa.ForeignKeyConstraint(['federal_register_id'], ['federal_register_entries.id'], ),
    sa.PrimaryKeyConstraint('id'),
    sa.UniqueConstraint('source_url')
    )
    op.create_index('idx_published_at_desc', 'articles', ['published_at'], unique=False)
    op.create_index('ix_articles_federal_register_id', 'articles', ['federal_register_id'], unique=False)
    op.create_index('ix_articles_published_at', 'articles', ['published_at'], unique=False)
    op.create_index('ix_articles_source_url', 'articles', ['source_url'], unique=False)


def downgrade() -> None:
    op.drop_index('ix_articles_source_url', table_name='articles')
    op.drop_index('ix_articles_published_at', table_name='articles')
    op.drop_index('ix_articles_federal_register_id', table_name='articles')
    op.drop_index('idx_published_at_desc', table_name='articles')
    op.drop_table('articles')
    op.drop_index('ix_federal_register_entries_processed', table_name='federal_register_entries')
    op.drop_index('ix_federal_register_entries_document_number', table_name='federal_register_entries')
    op.drop_index('idx_processed_fetched', table_name='federal_register_entries')
    op.drop_table('federal_register_entries')
