#!/usr/bin/env bash
set -e

cd "$(dirname "$0")/.."

if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"
DB_NAME="${DB_NAME:-opengov}"
DB_SSLMODE="${DB_SSLMODE:-prefer}"

export PGPASSWORD="$DB_PASSWORD"
export PGSSLMODE="$DB_SSLMODE"

run_query() {
    if [ $# -eq 0 ]; then
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"
    else
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$*"
    fi
}

reset_db() {
    echo "Terminating active connections to $DB_NAME..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();" 2>/dev/null || true

    echo "Dropping database $DB_NAME..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null || true

    echo "Creating database $DB_NAME..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;"

    echo "Running migrations..."
    go run ./cmd/scraper/main.go --migrate-only

    echo "Database $DB_NAME recreated successfully."
}

case "$1" in
    query)
        shift
        run_query "$@"
        ;;
    reset)
        reset_db
        ;;
    "")
        run_query
        ;;
    *)
        run_query "$@"
        ;;
esac
