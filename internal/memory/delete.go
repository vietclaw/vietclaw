package memory

import (
	"context"
	"fmt"
)

func (s *Store) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("memory id is required")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM memories WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if s.hasFTS {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM memories_fts WHERE rowid = ?`, id)
	}
	return nil
}
