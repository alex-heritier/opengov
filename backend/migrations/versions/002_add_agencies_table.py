"""Add agencies table

Revision ID: 002_add_agencies_table
Revises: 001_initial_schema
Create Date: 2025-11-13 19:30:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import sqlite

# revision identifiers, used by Alembic.
revision: str = '002_add_agencies_table'
down_revision: Union[str, None] = '001_initial_schema'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create agencies table
    op.create_table('agencies',
    sa.Column('id', sa.Integer(), nullable=False),
    sa.Column('fr_agency_id', sa.Integer(), nullable=False),
    sa.Column('name', sa.String(length=500), nullable=False),
    sa.Column('short_name', sa.String(length=200), nullable=True),
    sa.Column('slug', sa.String(length=200), nullable=False),
    sa.Column('description', sa.Text(), nullable=True),
    sa.Column('url', sa.String(length=500), nullable=True),
    sa.Column('json_url', sa.String(length=500), nullable=True),
    sa.Column('parent_id', sa.Integer(), nullable=True),
    sa.Column('raw_data', sqlite.JSON(), nullable=False),
    sa.Column('created_at', sa.DateTime(), nullable=False),
    sa.Column('updated_at', sa.DateTime(), nullable=False),
    sa.PrimaryKeyConstraint('id'),
    sa.UniqueConstraint('fr_agency_id'),
    sa.UniqueConstraint('slug')
    )
    op.create_index('idx_agency_name', 'agencies', ['name'], unique=False)
    op.create_index('ix_agencies_fr_agency_id', 'agencies', ['fr_agency_id'], unique=False)
    op.create_index('ix_agencies_id', 'agencies', ['id'], unique=False)
    op.create_index('ix_agencies_slug', 'agencies', ['slug'], unique=False)


def downgrade() -> None:
    op.drop_index('ix_agencies_slug', table_name='agencies')
    op.drop_index('ix_agencies_id', table_name='agencies')
    op.drop_index('ix_agencies_fr_agency_id', table_name='agencies')
    op.drop_index('idx_agency_name', table_name='agencies')
    op.drop_table('agencies')
