package infrastructure

import (
	"github.com/ivartj/kartotek/sqlmigrate/core"
	"github.com/ivartj/kartotek/sqlmigrate/core/entity"
	"github.com/ivartj/kartotek/util/sqlutil"
)

type SqliteMigrationStore struct {
	db core.DB
}

func NewSqliteMigrationStore(db core.DB) *SqliteMigrationStore {
	return &SqliteMigrationStore{
		db: db,
	}
}

func (store *SqliteMigrationStore) ListAll() ([]*entity.Migration, error) {
	rows, err := store.db.Query("select * from schema_migration")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrations := make([]*entity.Migration, 0)
	for rows.Next() {
		mig := new(entity.Migration)
		err = sqlutil.Rows{rows}.ScanEntity("", mig)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, mig)
	}
	return migrations, nil
}

func (store *SqliteMigrationStore) Register(mig *entity.Migration) error {
	return sqlutil.DB{store.db}.InsertOrReplaceEntity("schema_migration", mig)
}
