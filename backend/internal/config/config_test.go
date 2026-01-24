package config

import "testing"

func TestDatabaseURL_EncodesPassword(t *testing.T) {
	cfg := &Config{
		DatabaseHost: "localhost",
		DatabasePort: "5432",
		DatabaseUser: "user",
		DatabasePass: "my@pa:ss%word",
		DatabaseName: "opengov",
		DatabaseSSL:  "disable",
	}

	got := cfg.DatabaseURL()
	want := "postgres://user:my%40pa%3Ass%25word@localhost:5432/opengov?sslmode=disable"
	if got != want {
		t.Fatalf("DatabaseURL() = %q, want %q", got, want)
	}
}

func TestDatabaseURL_NoPassword(t *testing.T) {
	cfg := &Config{
		DatabaseHost: "localhost",
		DatabasePort: "5432",
		DatabaseUser: "user",
		DatabaseName: "opengov",
		DatabaseSSL:  "disable",
	}

	got := cfg.DatabaseURL()
	want := "postgres://user@localhost:5432/opengov?sslmode=disable"
	if got != want {
		t.Fatalf("DatabaseURL() = %q, want %q", got, want)
	}
}
