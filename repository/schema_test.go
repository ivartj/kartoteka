package repository

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func TestInitSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	err = InitSchema(db)
	if err != nil {
		t.Fatalf("Failed to initialize schema: %s", err)
	}
}
