package infrastructure

import (
	"database/sql"
	"github.com/ivartj/kartotek/sqlmigrate/core/entity"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrationLogGetLatest(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("PRAGMA foreign_keys = enable;")
	if err != nil {
		panic(err)
	}
	err = SqliteInitBaseSchema(db)
	if err != nil {
		panic(err)
	}
	store := NewSqliteMigrationStore(db)
	for _, mig := range []*entity.Migration{
		&entity.Migration{FromSchema: "ivartj.1", ToSchema: "ivartj.2"},
		&entity.Migration{FromSchema: "ivartj.2", ToSchema: "ivartj.3"},
	} {
		err = store.Register(mig)
		if err != nil {
			panic(err)
		}
	}
	log := NewSqliteMigrationLog(db)
	err = log.Add("ivartj.1", "ivartj.2")
	if err != nil {
		t.Fatalf("Failed to add log entry: %s", err)
	}
	err = log.Add("ivartj.2", "ivartj.3")
	if err != nil {
		t.Fatalf("Failed to add log entry: %s", err)
	}

	entry, err := log.GetLatest()
	if err != nil {
		t.Fatalf("Failed to get the latest log entry: %s", err)
	}
	assert.Equal(t, "ivartj.2", entry.FromSchema)
	assert.Equal(t, "ivartj.3", entry.ToSchema)
}
