package sqlutil

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Row struct {
	RowInterface
}

type RowInterface interface {
	Scan(dest ...interface{}) error
}

type Rows struct {
	RowsInterface
}

type RowsInterface interface {
	Scan(dest ...interface{}) error
	ColumnTypes() ([]*sql.ColumnType, error)
}

type DB struct {
	DBInterface
}

type DBInterface interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type entityField struct {
	sqlName    string
	structName string
	typ        reflect.Type
}

func getEntityTypeFields(typ reflect.Type) ([]*entityField, error) {
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("type is not a struct type")
	}
	entityFields := make([]*entityField, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		structField := typ.Field(i)
		sqlName, ok := structField.Tag.Lookup("sqlname")
		if !ok {
			continue
		}
		entityFields = append(entityFields, &entityField{sqlName, structField.Name, structField.Type})
	}
	return entityFields, nil
}

func getEntityTypeFieldMap(fields []*entityField) map[string]*entityField {
	m := map[string]*entityField{}
	for _, field := range fields {
		m[field.sqlName] = field
	}
	return m
}

func GetEntityTypeColumnString(prefix string, typ reflect.Type) (string, error) {
	fields, err := getEntityTypeFields(typ)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for i, f := range fields {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(prefix)
		b.WriteString(f.sqlName)
	}
	return b.String(), nil
}

func (row Row) ScanEntity(entity interface{}) error {
	value := reflect.ValueOf(entity)
	if value.Kind() != reflect.Ptr {
		return errors.New("entity parameter is not a pointer type")
	}
	value = value.Elem()
	fields, err := getEntityTypeFields(value.Type())
	if err != nil {
		return err
	}
	scanParameters := make([]reflect.Value, len(fields))
	for i, v := range fields {
		fieldValuePtr := value.FieldByName(v.structName).Addr()
		if fieldValuePtr.Type().Implements(sqlScannerType) {
			fieldValuePtr = fieldValuePtr.Convert(sqlScannerType)
		}
		scanParameters[i] = fieldValuePtr
	}
	retvals := reflect.ValueOf(row).MethodByName("Scan").Call(scanParameters)
	if !retvals[0].IsNil() {
		err, _ = retvals[0].Interface().(error)
		return err
	}
	return nil
}

var sqlScannerType reflect.Type = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

func (rows Rows) ScanEntity(columnPrefix string, entity interface{}) error {
	entityValue := reflect.ValueOf(entity)
	if entityValue.Kind() != reflect.Ptr {
		return errors.New("entity parameter is not a pointer type")
	}
	entityValue = entityValue.Elem()
	fields, err := getEntityTypeFields(entityValue.Type())
	if err != nil {
		return err
	}
	fieldMap := getEntityTypeFieldMap(fields)
	m := map[string]interface{}{}
	err = rows.ScanMap(m)
	if err != nil {
		return err
	}
	for columnName, value := range m {
		if reflect.TypeOf(value) == nullType {
			continue
		}
		if !strings.HasPrefix(columnName, columnPrefix) {
			continue
		}
		fieldName := strings.TrimPrefix(columnName, columnPrefix)
		field, ok := fieldMap[fieldName]
		if ok {
			fieldValue := entityValue.FieldByName(field.structName)
			if fieldValue.Addr().Type().Implements(sqlScannerType) {
				fieldValue = fieldValue.Addr().Convert(sqlScannerType)
				retValues := fieldValue.MethodByName("Scan").Call([]reflect.Value{reflect.ValueOf(value)})
				if !retValues[0].IsNil() {
					return retValues[0].Interface().(error)
				}
			} else {
				fieldValue.Set(reflect.ValueOf(value).Convert(field.typ))
			}
		}
	}
	return nil
}

type null struct{}

var nullType reflect.Type = reflect.TypeOf(null{})

func (n null) Value() (driver.Value, error) {
	return nil, nil
}

func (n *null) Scan(value interface{}) error {
	if value != nil {
		return fmt.Errorf("%s is not nil", value)
	}
	return nil
}

func (rows Rows) ScanMap(m map[string]interface{}) error {
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	values := make([]reflect.Value, len(columnTypes))
	for i, column := range columnTypes {
		if column.ScanType() == nil {
			values[i] = reflect.ValueOf(sql.Scanner(&null{}))
		} else {
			values[i] = reflect.New(column.ScanType())
		}
		// values[i] = reflect.New(column.ScanType()).Addr()
	}
	retValues := reflect.ValueOf(rows.Scan).Call(values)
	if !retValues[0].IsNil() {
		err, _ = retValues[0].Interface().(error)
		return err
	}
	for i, column := range columnTypes {
		m[column.Name()] = values[i].Elem().Interface()
	}
	return nil
}

func (db DB) InsertEntity(table string, entity interface{}) error {
	return db.insertEntity(table, entity, false)
}

func (db DB) InsertOrReplaceEntity(table string, entity interface{}) error {
	return db.insertEntity(table, entity, true)
}

var sqlValuerType reflect.Type = reflect.ValueOf((*driver.Valuer)(nil)).Type().Elem()

func (db DB) insertEntity(table string, entity interface{}, orReplace bool) error {
	entityValue := reflect.ValueOf(entity)
	if entityValue.Kind() != reflect.Ptr {
		return errors.New("entity parameter is not a pointer type")
	}
	entityValue = entityValue.Elem()
	entityType := entityValue.Type()
	entityFields, err := getEntityTypeFields(entityType)
	if err != nil {
		return err
	}
	columnString, err := GetEntityTypeColumnString("", entityType)
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString("INSERT ")
	if orReplace {
		sb.WriteString("OR REPLACE ")
	}
	sb.WriteString("INTO ")
	sb.WriteString(table)
	sb.WriteString(" (")
	sb.WriteString(columnString)
	sb.WriteString(") VALUES (")
	for i := range entityFields {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("?")
	}
	sb.WriteString(");")

	execParams := make([]reflect.Value, 1+len(entityFields))
	execParams[0] = reflect.ValueOf(sb.String())
	for i, entityField := range entityFields {
		value := entityValue.FieldByName(entityField.structName)
		if value.Type().Implements(sqlValuerType) {
			value = value.Convert(sqlValuerType)
		}
		execParams[i+1] = value
	}
	retValues := reflect.ValueOf(db.Exec).Call(execParams)
	if !retValues[1].IsNil() {
		err, _ = retValues[1].Interface().(error)
		return err
	}
	// TODO: Check sql.Result return value that an insert occurred
	return nil
}
