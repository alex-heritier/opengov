package migration

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
)

//go:embed *.sql
var migrationsFS embed.FS

var migrationNameRE = regexp.MustCompile(`^\d{3}_.+\.sql$`)

func FS() fs.FS {
	return migrationsFS
}

func List() ([]string, error) {
	files, err := fs.Glob(migrationsFS, "*.sql")
	if err != nil {
		return nil, fmt.Errorf("failed to glob migration files: %w", err)
	}
	sort.Strings(files)

	for _, f := range files {
		if !migrationNameRE.MatchString(f) {
			return nil, fmt.Errorf("invalid migration filename %q (expected NNN_description.sql)", f)
		}
	}

	return files, nil
}

func ReadFile(name string) ([]byte, error) {
	b, err := migrationsFS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration %q: %w", name, err)
	}
	return b, nil
}
