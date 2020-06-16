package infrastructure

import (
	"database/sql"
	"github.com/ivartj/kartotek/sqlmigrate/core"
	"github.com/ivartj/kartotek/sqlmigrate/core/entity"
	"github.com/ivartj/kartotek/util/sqlutil"
	"time"
)

type SqliteMigrationLog struct {
	db core.DB
}

var logEntryColumnString string

func NewSqliteMigrationLog(db core.DB) *SqliteMigrationLog {
	return &SqliteMigrationLog{
		db: db,
	}
}

func (log *SqliteMigrationLog) Add(from, to string) error {
	utcTime := time.Now().UTC()
	entry := entity.MigrationLogEntry{
		UtcTime:    &utcTime,
		FromSchema: from,
		ToSchema:   to,
	}
	return sqlutil.DB{log.db}.InsertEntity("schema_log", &entry)
}

func (log *SqliteMigrationLog) GetLatest() (*entity.MigrationLogEntry, error) {
	row := log.db.QueryRow("select * from schema_log order by utc_time desc;")
	var entry entity.MigrationLogEntry
	err := sqlutil.Row{row}.ScanEntity(&entry)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entry, nil
}
