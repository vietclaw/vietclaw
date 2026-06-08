package memory

import (
	"context"
	"fmt"
	"strings"
)

func (s *Store) List(ctx context.Context, scope string, limit int) ([]Record, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	args := []any{}
	query := `SELECT id, scope, kind, content, confidence, created_at, updated_at FROM memories`
	if scope != "" {
		query += ` WHERE scope = ?`
		args = append(args, scope)
	}
	query += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list memories: %w", err)
	}
	defer rows.Close()
	return scanRecords(rows)
}

func (s *Store) Search(ctx context.Context, scope, query string, limit int) ([]Record, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.List(ctx, scope, limit)
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if s.hasFTS {
		records, err := s.searchFTS(ctx, scope, query, limit)
		if err == nil {
			return records, nil
		}
	}
	return s.searchLike(ctx, scope, query, limit)
}

func (s *Store) searchFTS(ctx context.Context, scope, query string, limit int) ([]Record, error) {
	args := []any{query}
	sqlQuery := `
SELECT m.id, m.scope, m.kind, m.content, m.confidence, m.created_at, m.updated_at
FROM memories_fts f
JOIN memories m ON m.id = f.rowid
WHERE memories_fts MATCH ?`
	if scope != "" {
		sqlQuery += ` AND m.scope = ?`
		args = append(args, scope)
	}
	sqlQuery += ` ORDER BY m.id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRecords(rows)
}

func (s *Store) searchLike(ctx context.Context, scope, query string, limit int) ([]Record, error) {
	like := "%" + strings.ToLower(query) + "%"
	args := []any{like}
	sqlQuery := `
SELECT id, scope, kind, content, confidence, created_at, updated_at
FROM memories
WHERE lower(content) LIKE ?`
	if scope != "" {
		sqlQuery += ` AND scope = ?`
		args = append(args, scope)
	}
	sqlQuery += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("search memories: %w", err)
	}
	defer rows.Close()
	return scanRecords(rows)
}
