package infrastructure

import (
	"fmt"
	"github.com/ivartj/kartoteka/sqlmigrate/core"
)

func SqliteInitBaseSchema(db core.DB) error {
	_, err := db.Exec(`
		create table if not exists schema_migration (
			from_schema text not null,
			to_schema text not null,
			sql_code text not null,
			primary key(from_schema, to_schema)
		);`)
	if err != nil {
		return fmt.Errorf("Failed to initialize base SQLite schema: %w", err)
	}

	_, err = db.Exec(`
		create table if not exists schema_log (
			utc_time datetime not null,
			from_schema text not null,
			to_schema text not null,
			foreign key(from_schema, to_schema)
			  references schema_migration(from_schema, to_schema)
		);
	`)
	if err != nil {
		return fmt.Errorf("Failed to initialize base SQLite schema: %w", err)
	}
	return nil
}
