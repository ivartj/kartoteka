package entities

import (
	"database/sql"
	"reflect"
	"testing"
)

var sqlScannerType reflect.Type = reflect.ValueOf((*sql.Scanner)(nil)).Type().Elem()

func TestIDIsScannable(t *testing.T) {
	id := UserID(NewID())
	test(&id)
	if !reflect.ValueOf(&id).Type().Implements(sqlScannerType) {
		t.Errorf("ID instance's dynamic type does not implement sql.Scanner")
	}
}

func test(sc sql.Scanner) {

}
