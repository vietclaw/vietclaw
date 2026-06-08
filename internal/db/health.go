package db

import "database/sql"

func Check(database *sql.DB) bool {
	if database == nil {
		return false
	}
	return database.Ping() == nil
}
