package sqlutil

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type testEntity struct {
	FieldWithoutCorrespondingSQLColumn int
	ID                                 int    `sqlname:"user_id"`
	Name                               string `sqlname:"user_name"`
}

func TestGetEntityColumnString(t *testing.T) {
	typ := reflect.TypeOf(testEntity{})
	str, err := GetEntityTypeColumnString("", typ)
	if err != nil {
		t.Fatalf("%s", err)
	}
	assert.Equal(t, "user_id, user_name", str)
}

type testRow struct{}

func (sc testRow) Scan(dest ...interface{}) error {
	for i, v := range dest {
		switch i {
		case 0:
			id, ok := v.(*int)
			if !ok {
				return fmt.Errorf("First argument to Scan not a pointer to int, but %s", reflect.TypeOf(v).Name())
			}
			*id = 2384
		case 1:
			name, ok := v.(*string)
			if !ok {
				return errors.New("Second argument to Scan not a pointer to string")
			}
			*name = "ivartj"
		default:
			return errors.New("More than expected number of arguments to Scan")
		}
	}
	return nil
}

func TestRowScanEntity(t *testing.T) {
	row := Row{testRow{}}
	var entity testEntity
	err := row.ScanEntity(&entity)
	if err != nil {
		t.Fatalf("ScanEntity: %s", err)
	}
	assert.Equal(t, 2384, entity.ID)
	assert.Equal(t, "ivartj", entity.Name)
}

func TestRowsScanMap(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	rows, err := db.Query("select 123 as user_id, 'ivartj' as user_name;")
	if err != nil {
		panic(err)
	}
	ok := rows.Next()
	if !ok {
		panic("No row")
	}
	m := map[string]interface{}{}
	err = Rows{rows}.ScanMap(m)
	if err != nil {
		t.Fatalf("Failed to scan map: %s", err)
	}
	assert.Equal(t, int64(123), m["user_id"])
	assert.Equal(t, "ivartj", m["user_name"])
}

func TestRowsScanEntity(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	rows, err := db.Query("select 123 as user_id, 'ivartj' as user_name;")
	if err != nil {
		panic(err)
	}
	ok := rows.Next()
	if !ok {
		panic("No row")
	}
	var entity testEntity
	err = Rows{rows}.ScanEntity("", &entity)
	if err != nil {
		t.Fatalf("Failed to scan entity: %s", err)
	}
	assert.Equal(t, 123, entity.ID)
	assert.Equal(t, "ivartj", entity.Name)
}

func TestDbInsertEntity(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`
		create table user (
			user_id integer not null primary key,
			user_name text not null
		);
	`)
	if err != nil {
		panic(err)
	}

	{
		user := &testEntity{
			ID:   123,
			Name: "ivartj",
		}
		err = DB{db}.InsertEntity("user", user)
		if err != nil {
			t.Fatalf("Error on inserting entity: %s", err)
		}
	}
	row := db.QueryRow("select user_id, user_name from user;")
	var user testEntity
	err = Row{row}.ScanEntity(&user)
	if err != nil {
		t.Fatalf("Error scanning inserted entity: %s", err)
	}

	assert.Equal(t, 123, user.ID)
	assert.Equal(t, "ivartj", user.Name)
}
