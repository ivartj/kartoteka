package entity

import (
	"time"
)

type MigrationLogEntry struct {
	UtcTime    *time.Time `sqlname:"utc_time"`
	FromSchema string     `sqlname:"from_schema"`
	ToSchema   string     `sqlname:"to_schema"`
}
