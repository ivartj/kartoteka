package sqlmigrate

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

type testMigration struct {
	from, to, sql string
}

var testMigrations = []testMigration{
	testMigration{
		from: "",
		to:   "ivartj.1",
		sql:  `create table user ( id integer primary key, name text not null );`,
	},
	testMigration{
		from: "ivartj.1",
		to:   "ivartj.2",
		sql:  `alter table user add column email text not null;`,
	},
	testMigration{
		from: "",
		to:   "ivartj.2",
		sql:  `create table user ( id integer primary key, name text not null, email text not null );`,
	},
	testMigration{
		from: "ivartj.2",
		to:   "ivartj.3",
		sql: `
			create table role (
				id integer primary key,
				name text not null
			);
			create table user_grp (
				user_id integer not null
					references user(id),
				role_id integer not null
					references role(id)
			);
		`,
	},
}

func prep() (*M, *sql.DB) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("pragma foreign_keys = on")
	if err != nil {
		panic(err)
	}
	m, err := New(db)
	if err != nil {
		panic(err)
	}
	return m, db
}

func assertTableExists(t *testing.T, db *sql.DB, tableName string) {
	row := db.QueryRow("select name from sqlite_master where type = 'table' and name = '" + tableName + "';")
	err := row.Scan(&tableName)
	if err != nil {
		t.Fatalf("Failed to scan expected table: %s", err)
	}
}

func TestMigrateTo(t *testing.T) {
	m, db := prep()
	var err error
	for _, mig := range testMigrations {
		err = m.RegisterMigration(mig.from, mig.to, mig.sql)
		if err != nil {
			t.Fatalf("Failed to register migration: %s", err)
		}
	}
	err = m.MigrateTo("ivartj.2")
	if err != nil {
		t.Fatalf("Failed to migrate: %s", err)
	}
	assertTableExists(t, db, "user")
	err = m.MigrateTo("ivartj.3")
	if err != nil {
		t.Fatalf("Failed to migrate: %s", err)
	}
	assertTableExists(t, db, "user_grp")
}
