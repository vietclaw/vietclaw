package memory

import (
	"context"
	"strings"
)

type CurationResult struct {
	Removed int64 `json:"removed"`
}

func (s *Store) CurateDuplicates(ctx context.Context, scope string) (CurationResult, error) {
	records, err := s.List(ctx, scope, 200)
	if err != nil {
		return CurationResult{}, err
	}
	seen := map[string]int64{}
	var removed int64
	for _, rec := range records {
		key := rec.Scope + "\x00" + strings.ToLower(strings.Join(strings.Fields(rec.Content), " "))
		if key == rec.Scope+"\x00" {
			continue
		}
		if _, ok := seen[key]; ok {
			if err := s.Delete(ctx, rec.ID); err != nil {
				return CurationResult{}, err
			}
			removed++
			continue
		}
		seen[key] = rec.ID
	}
	return CurationResult{Removed: removed}, nil
}
