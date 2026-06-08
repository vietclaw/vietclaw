package memory

import "database/sql"

func scanRecords(rows *sql.Rows) ([]Record, error) {
	records := []Record{}
	for rows.Next() {
		var rec Record
		var kind string
		var confidence float64
		if err := rows.Scan(&rec.ID, &rec.Scope, &kind, &rec.Content, &confidence, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		rec.Kind = Kind(kind)
		rec.Confidence = confidenceLabel(confidence)
		records = append(records, rec)
	}
	return records, rows.Err()
}
