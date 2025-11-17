"""Add fastapi-users required fields

Revision ID: 004_add_fastapi_users_fields
Revises: 003_add_users_table
Create Date: 2025-11-17 00:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '004_add_fastapi_users_fields'
down_revision: Union[str, None] = '003_add_users_table'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Add hashed_password column (required by fastapi-users)
    # SQLite doesn't support ALTER COLUMN, so we add it with a server_default
    op.add_column('users', sa.Column('hashed_password', sa.String(length=1024), nullable=False, server_default=''))

    # Add is_superuser column (required by fastapi-users)
    op.add_column('users', sa.Column('is_superuser', sa.Boolean(), nullable=False, server_default='0'))

    # Note: In SQLite, we cannot easily remove server_default after creation
    # This is acceptable for this migration as the defaults are safe


def downgrade() -> None:
    op.drop_column('users', 'is_superuser')
    op.drop_column('users', 'hashed_password')
