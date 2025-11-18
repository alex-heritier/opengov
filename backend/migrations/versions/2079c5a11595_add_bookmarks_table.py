"""add bookmarks table

Revision ID: 2079c5a11595
Revises: 004_merge_article_federal_register
Create Date: 2025-11-17 19:04:54.559764

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '2079c5a11595'
down_revision: Union[str, None] = '004_merge_article_federal_register'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create bookmarks table
    op.create_table('bookmarks',
        sa.Column('id', sa.Integer(), nullable=False),
        sa.Column('user_id', sa.Integer(), nullable=False),
        sa.Column('frarticle_id', sa.Integer(), nullable=False),
        sa.Column('is_bookmarked', sa.Boolean(), nullable=False, server_default='1'),
        sa.Column('created_at', sa.DateTime(), nullable=False),
        sa.Column('updated_at', sa.DateTime(), nullable=False),

        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['frarticle_id'], ['frarticles.id'], ondelete='CASCADE'),
        sa.UniqueConstraint('user_id', 'frarticle_id', name='uix_user_article_bookmark')
    )

    # Create indexes
    op.create_index('ix_bookmarks_id', 'bookmarks', ['id'], unique=False)
    op.create_index('idx_bookmarks_user_id', 'bookmarks', ['user_id'], unique=False)
    op.create_index('idx_bookmarks_frarticle_id', 'bookmarks', ['frarticle_id'], unique=False)
    op.create_index('idx_bookmarks_user_bookmarked', 'bookmarks', ['user_id', 'is_bookmarked'], unique=False)


def downgrade() -> None:
    op.drop_index('idx_bookmarks_user_bookmarked', table_name='bookmarks')
    op.drop_index('idx_bookmarks_frarticle_id', table_name='bookmarks')
    op.drop_index('idx_bookmarks_user_id', table_name='bookmarks')
    op.drop_index('ix_bookmarks_id', table_name='bookmarks')
    op.drop_table('bookmarks')
