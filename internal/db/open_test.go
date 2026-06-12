package db

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestOpen_ConcurrentWrites(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "concurrent.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer database.Close()

	if err := ApplySchema(database); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	const workers = 8
	const writesPerWorker = 40

	g, ctx := errgroup.WithContext(context.Background())
	for worker := 0; worker < workers; worker++ {
		worker := worker
		g.Go(func() error {
			sessionID := fmt.Sprintf("parent:spawn:agent-%d:sub_abc:delegate:agent-%d", worker, worker)
			for i := 0; i < writesPerWorker; i++ {
				now := time.Now().UTC().Format(time.RFC3339)
				if _, err := database.ExecContext(ctx, `
INSERT INTO sessions (id, channel, user_id, title, summary, created_at, updated_at)
VALUES (?, 'web', 'test', NULL, NULL, ?, ?)
ON CONFLICT(id) DO UPDATE SET updated_at = excluded.updated_at`,
					sessionID, now, now); err != nil {
					return err
				}
				if _, err := database.ExecContext(ctx, `
INSERT INTO messages (session_id, role, content, created_at)
VALUES (?, 'user', ?, ?)`,
					sessionID, fmt.Sprintf("task-%d-%d", worker, i), now); err != nil {
					return err
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Fatalf("concurrent writes failed: %v", err)
	}

	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&count); err != nil {
		t.Fatalf("count messages: %v", err)
	}
	want := workers * writesPerWorker
	if count != want {
		t.Fatalf("message count = %d, want %d", count, want)
	}
}

func TestOpen_WALMode(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "wal.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer database.Close()

	var journalMode string
	if err := database.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("read journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", journalMode)
	}
}

func TestOpen_SerializesWithoutBusyErrors(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "serial.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer database.Close()

	if err := ApplySchema(database); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tx, err := database.Begin()
			if err != nil {
				errCh <- err
				return
			}
			now := time.Now().UTC().Format(time.RFC3339)
			_, err = tx.Exec(`
INSERT INTO sessions (id, channel, user_id, title, summary, created_at, updated_at)
VALUES (?, 'web', 'test', NULL, NULL, ?, ?)`,
				fmt.Sprintf("session-%d", i), now, now)
			if err != nil {
				_ = tx.Rollback()
				errCh <- err
				return
			}
			if err := tx.Commit(); err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("transaction failed: %v", err)
		}
	}
}
