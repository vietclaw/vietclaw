package memory

import (
	"context"
	"database/sql"
)

type Store struct {
	db     *sql.DB
	hasFTS bool
}

func NewStore(db *sql.DB) *Store {
	store := &Store{db: db}
	store.hasFTS = store.ensureFTS(context.Background()) == nil
	return store
}

func (s *Store) ensureFTS(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(scope, kind, content)`)
	return err
}
