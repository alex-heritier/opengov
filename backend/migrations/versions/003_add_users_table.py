"""Add users table with fastapi-users support

Revision ID: 003_add_users_table
Revises: 002_add_agencies_table
Create Date: 2025-11-17 00:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '003_add_users_table'
down_revision: Union[str, None] = '002_add_agencies_table'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create users table with all fastapi-users required fields
    op.create_table('users',
        sa.Column('id', sa.Integer(), nullable=False),
        sa.Column('email', sa.String(length=255), nullable=False),
        sa.Column('hashed_password', sa.String(length=1024), nullable=False),
        sa.Column('is_active', sa.Boolean(), nullable=False, server_default='1'),
        sa.Column('is_superuser', sa.Boolean(), nullable=False, server_default='0'),
        sa.Column('is_verified', sa.Boolean(), nullable=False, server_default='0'),

        # Optional Google OAuth fields
        sa.Column('google_id', sa.String(length=255), nullable=True),
        sa.Column('name', sa.String(length=255), nullable=True),
        sa.Column('picture_url', sa.String(length=500), nullable=True),

        # Timestamps
        sa.Column('created_at', sa.DateTime(), nullable=False),
        sa.Column('updated_at', sa.DateTime(), nullable=False),
        sa.Column('last_login_at', sa.DateTime(), nullable=True),

        sa.PrimaryKeyConstraint('id')
    )

    # Create indexes
    op.create_index('ix_users_id', 'users', ['id'], unique=False)
    op.create_index('ix_users_email', 'users', ['email'], unique=True)
    op.create_index('ix_users_google_id', 'users', ['google_id'], unique=True)


def downgrade() -> None:
    op.drop_index('ix_users_google_id', table_name='users')
    op.drop_index('ix_users_email', table_name='users')
    op.drop_index('ix_users_id', table_name='users')
    op.drop_table('users')
