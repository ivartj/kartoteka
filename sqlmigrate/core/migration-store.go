package core

import (
	"github.com/ivartj/kartotek/sqlmigrate/core/entity"
)

type MigrationStore interface {
	Register(migration *entity.Migration) error
	ListAll() ([]*entity.Migration, error)
}
