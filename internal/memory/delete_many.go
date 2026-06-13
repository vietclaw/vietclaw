package memory

import (
	"context"
	"fmt"
	"strings"
)

// DeleteMany deletes multiple memory records by their IDs.
func (s *Store) DeleteMany(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	// For simple list of IDs, we can use the IN clause.
	// Since sqlite limits the number of variables in a statement,
	// we may need to batch this if there are too many IDs.
	// SQLite max variable limit is generally 32766 or 999.
	// 500 is a safe batch size.
	const batchSize = 500

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Will be ignored if tx is committed

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}

		batch := ids[i:end]

		placeholders := make([]string, len(batch))
		args := make([]interface{}, len(batch))
		for j, id := range batch {
			placeholders[j] = "?"
			args[j] = id
		}

		query := fmt.Sprintf("DELETE FROM memories WHERE id IN (%s)", strings.Join(placeholders, ","))

		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}

		if s.hasFTS {
			ftsQuery := fmt.Sprintf("DELETE FROM memories_fts WHERE rowid IN (%s)", strings.Join(placeholders, ","))
			_, err = tx.ExecContext(ctx, ftsQuery, args...)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
