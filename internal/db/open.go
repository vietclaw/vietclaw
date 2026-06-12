package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const busyTimeoutMS = 15000

func dsn(path string) string {
	slash := filepath.ToSlash(path)
	if strings.Contains(slash, "?") {
		slash = strings.ReplaceAll(slash, "?", "%3F")
	}
	pragma := fmt.Sprintf(
		"_pragma=busy_timeout(%d)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)",
		busyTimeoutMS,
	)
	return "file:" + slash + "?" + pragma
}

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database dir: %w", err)
	}

	database, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// WAL allows concurrent readers with serialized writers; busy_timeout retries on lock.
	database.SetMaxOpenConns(8)

	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := configureSQLite(database); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("configure database: %w", err)
	}

	return database, nil
}

func configureSQLite(database *sql.DB) error {
	if _, err := database.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d", busyTimeoutMS)); err != nil {
		return err
	}
	var journalMode string
	if err := database.QueryRow("PRAGMA journal_mode=WAL").Scan(&journalMode); err != nil {
		return err
	}
	return nil
}
