package entity

type Migration struct {
	FromSchema string `sqlname:"from_schema"`
	ToSchema   string `sqlname:"to_schema"`
	SqlCode    string `sqlname:"sql_code"`
}
