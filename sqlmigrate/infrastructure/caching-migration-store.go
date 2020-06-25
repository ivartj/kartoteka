package infrastructure

import (
	"encoding/hex"
	"github.com/ivartj/kartoteka/sqlmigrate/core"
	"github.com/ivartj/kartoteka/sqlmigrate/core/entity"
	"strings"
)

type CachingMigrationStore struct {
	dbStore core.MigrationStore
	migMap  map[string]*entity.Migration
	list    []*entity.Migration
}

func NewCachingMigrationStore(dbStore core.MigrationStore) *CachingMigrationStore {
	return &CachingMigrationStore{
		dbStore: dbStore,
		migMap:  map[string]*entity.Migration{},
	}
}

func (store *CachingMigrationStore) ListAll() ([]*entity.Migration, error) {
	if store.list != nil {
		return store.list, nil
	}
	list, err := store.dbStore.ListAll()
	if err != nil {
		return nil, err
	}
	store.list = list
	for _, mig := range list {
		store.migMap[migMapId(mig)] = mig
	}
	return list, nil
}

func (store *CachingMigrationStore) Register(mig *entity.Migration) error {
	cached, ok := store.migMap[migMapId(mig)]
	if ok && cached.SqlCode == mig.SqlCode {
		return nil
	}
	err := store.dbStore.Register(mig)
	if err != nil {
		return err
	}
	if ok {
		store.list = nil
	}
	store.migMap[migMapId(mig)] = mig

	return nil
}

func migMapId(mig *entity.Migration) string {
	var sb strings.Builder
	sb.WriteString(hex.EncodeToString([]byte(mig.FromSchema)))
	sb.WriteString(":")
	sb.WriteString(hex.EncodeToString([]byte(mig.ToSchema)))
	return sb.String()
}
