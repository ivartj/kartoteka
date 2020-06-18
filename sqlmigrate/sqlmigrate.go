package sqlmigrate

import (
	"github.com/ivartj/kartotek/sqlmigrate/core"
	"github.com/ivartj/kartotek/sqlmigrate/core/entity"
	"github.com/ivartj/kartotek/sqlmigrate/infrastructure"
)

type M struct {
	db    core.DB
	store core.MigrationStore
	log   core.MigrationLog
}

func New(db core.DB) (*M, error) {
	m := &M{
		db:    db,
		store: infrastructure.NewCachingMigrationStore(infrastructure.NewSqliteMigrationStore(db)),
		log:   infrastructure.NewSqliteMigrationLog(db),
	}
	err := infrastructure.SqliteInitBaseSchema(m.db)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *M) MigrateTo(schema string) error {
	entry, err := m.log.GetLatest()
	if err != nil {
		return err
	}
	currentSchema := ""
	if entry != nil {
		currentSchema = entry.ToSchema
	}
	if currentSchema == schema {
		return nil
	}
	pathFinder := infrastructure.NewPathFinder(m.store)
	path, err := pathFinder.FindPath(currentSchema, schema)
	if err != nil {
		return err
	}
	for _, mig := range path {
		_, err = m.db.Exec(mig.SqlCode)
		if err != nil {
			return err
		}
		err = m.log.Add(mig.FromSchema, mig.ToSchema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *M) RegisterMigration(fromSchema, toSchema, sqlCode string) error {
	err := m.store.Register(&entity.Migration{
		FromSchema: fromSchema,
		ToSchema:   toSchema,
		SqlCode:    sqlCode,
	})
	if err != nil {
		return err
	}
	return nil
}
