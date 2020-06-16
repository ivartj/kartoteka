package infrastructure

import (
	"github.com/ivartj/kartotek/sqlmigrate/core/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockMigrationStore struct{}

func (store mockMigrationStore) Register(mig *entity.Migration) error {
	panic("unimplemented")
}

func (store mockMigrationStore) ListAll() ([]*entity.Migration, error) {
	migs := []*entity.Migration{
		&entity.Migration{FromSchema: "ivartj.1", ToSchema: "ivartj.2"},
		&entity.Migration{FromSchema: "ivartj.2", ToSchema: "ivartj.3"},
		&entity.Migration{FromSchema: "ivartj.3", ToSchema: "ivartj.4"},
		&entity.Migration{FromSchema: "ivartj.1", ToSchema: "ivartj.3"},
	}
	return migs, nil
}

func TestPathFinderFindPath(t *testing.T) {
	var store mockMigrationStore
	pathFinder := NewPathFinder(store)
	path, err := pathFinder.FindPath("ivartj.1", "ivartj.4")
	if err != nil {
		t.Fatalf("Failed to find path: %s", err)
	}
	assert.Equal(t, 2, len(path))
	assert.Equal(t, "ivartj.1", path[0].FromSchema)
	assert.Equal(t, "ivartj.3", path[0].ToSchema)
	assert.Equal(t, "ivartj.3", path[1].FromSchema)
	assert.Equal(t, "ivartj.4", path[1].ToSchema)
}
