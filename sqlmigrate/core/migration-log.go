package core

import (
	"github.com/ivartj/kartoteka/sqlmigrate/core/entity"
)

type MigrationLog interface {
	GetLatest() (*entity.MigrationLogEntry, error)
	Add(from, to string) error
}
