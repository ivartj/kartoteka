package infrastructure

import (
	"database/sql"
	"github.com/ivartj/kartoteka/sqlmigrate/core/entity"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_, err = db.Exec("PRAGMA foreign_keys = enable;")
	if err != nil {
		panic(err)
	}
	err = SqliteInitBaseSchema(db)
	if err != nil {
		panic(err)
	}

	store := NewSqliteMigrationStore(db)

	err = store.Register(&entity.Migration{
		FromSchema: "",
		ToSchema:   "0.1",
		SqlCode: `
			create table user (
				user_id integer not null
					primary key,
			  username text not null
					unique,
				password_hash not null
			);
		`})
	if err != nil {
		t.Fatalf("Failed to add a migration: %s", err)
	}

	migs, err := store.ListAll()
	if err != nil {
		t.Fatalf("Error on listing migrations: %s", err)
	}

	assert.Equal(t, 1, len(migs))
	assert.Equal(t, "", migs[0].FromSchema)
	assert.Equal(t, "0.1", migs[0].ToSchema)
}
