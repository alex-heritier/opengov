package db

import (
	"fmt"
	"strings"

	"github.com/alex/opengov-go/migration"
)

func (db *DB) RunMigrations() error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	files, err := migration.List()
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	for _, file := range files {
		content, err := migration.ReadFile(file)
		if err != nil {
			return err
		}

		statements := splitStatements(string(content))
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", file, err)
			}
		}
	}

	return tx.Commit()
}

func splitStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	inComment := false
	parens := 0

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		if inComment {
			if ch == '\n' {
				inComment = false
			}
			continue
		}

		if ch == '-' && i+1 < len(sql) && sql[i+1] == '-' {
			inComment = true
			i++
			continue
		}

		if ch == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		}
		if ch == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		}

		if !inSingleQuote && !inDoubleQuote {
			if ch == '(' {
				parens++
			}
			if ch == ')' {
				parens--
			}
		}

		if ch == ';' && parens == 0 && !inSingleQuote && !inDoubleQuote && !inComment {
			stmt := current.String()
			if strings.TrimSpace(stmt) != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}

	stmt := current.String()
	if strings.TrimSpace(stmt) != "" {
		statements = append(statements, stmt)
	}

	return statements
}
