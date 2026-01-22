package db

import (
	"fmt"
	"strings"
)

const MigrationSQL = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1,
    is_superuser INTEGER NOT NULL DEFAULT 0,
    is_verified INTEGER NOT NULL DEFAULT 0,
    google_id TEXT UNIQUE,
    name TEXT,
    picture_url TEXT,
    political_leaning TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    last_login_at TEXT
);

CREATE TABLE IF NOT EXISTS agencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    fr_agency_id INTEGER NOT NULL UNIQUE,
    name TEXT NOT NULL,
    short_name TEXT,
    slug TEXT NOT NULL,
    description TEXT,
    url TEXT,
    json_url TEXT,
    parent_id INTEGER,
    raw_data TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS frarticles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    unique_key TEXT NOT NULL UNIQUE,
    document_number TEXT NOT NULL,
    raw_data TEXT NOT NULL,
    fetched_at TEXT NOT NULL DEFAULT (datetime('now')),
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    source_url TEXT NOT NULL,
    published_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    frarticle_id INTEGER NOT NULL REFERENCES frarticles(id) ON DELETE CASCADE,
    is_bookmarked INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(user_id, frarticle_id)
);

CREATE TABLE IF NOT EXISTS likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    frarticle_id INTEGER NOT NULL REFERENCES frarticles(id) ON DELETE CASCADE,
    is_liked INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(user_id, frarticle_id)
);
`

func (db *DB) RunMigrations() error {
	statements := strings.Split(MigrationSQL, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || stmt == "PRAGMA foreign_keys = ON;" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	if err := db.addMissingColumns(); err != nil {
		return fmt.Errorf("failed to add missing columns: %w", err)
	}

	return nil
}

func (db *DB) addMissingColumns() error {
	var count int

	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('frarticles') WHERE name='source'").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = db.Exec("ALTER TABLE frarticles ADD COLUMN source TEXT NOT NULL DEFAULT 'fedreg'")
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}

		_, err = db.Exec("ALTER TABLE frarticles ADD COLUMN source_id TEXT NOT NULL DEFAULT ''")
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}

		_, err = db.Exec("ALTER TABLE frarticles ADD COLUMN unique_key TEXT")
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}

		_, err = db.Exec("UPDATE frarticles SET source = 'fedreg', source_id = document_number, unique_key = 'fedreg:' || document_number WHERE source IS NULL OR source = ''")
		if err != nil {
			return fmt.Errorf("failed to backfill source columns: %w", err)
		}

		_, err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_frarticles_unique_key ON frarticles(unique_key)")
		if err != nil && !strings.Contains(err.Error(), "index") {
			return fmt.Errorf("failed to create unique_key index: %w", err)
		}
	}

	return nil
}
