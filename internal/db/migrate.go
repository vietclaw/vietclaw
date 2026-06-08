package db

import (
	"database/sql"
	"fmt"
)

func ApplySchema(database *sql.DB) error {
	if _, err := database.Exec(schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
