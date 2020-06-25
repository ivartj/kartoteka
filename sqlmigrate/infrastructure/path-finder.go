package infrastructure

import (
	"container/list"
	"errors"
	"github.com/ivartj/kartoteka/sqlmigrate/core"
	"github.com/ivartj/kartoteka/sqlmigrate/core/entity"
)

type PathFinder struct {
	store core.MigrationStore
}

func NewPathFinder(store core.MigrationStore) *PathFinder {
	return &PathFinder{
		store: store,
	}
}

func (pf *PathFinder) FindPath(from, to string) ([]*entity.Migration, error) {
	migs, err := pf.store.ListAll()
	if err != nil {
		return nil, err
	}

	toFromMap := map[string][]string{} // key is destination, value is possible origins
	migMap := map[string]map[string]*entity.Migration{}
	for _, mig := range migs {
		destinations, ok := toFromMap[mig.ToSchema]
		if !ok {
			toFromMap[mig.ToSchema] = []string{mig.FromSchema}
			migMap[mig.ToSchema] = map[string]*entity.Migration{mig.FromSchema: mig}
		} else {
			toFromMap[mig.ToSchema] = append(destinations, mig.FromSchema)
			migMap[mig.ToSchema][mig.FromSchema] = mig
		}
	}

	pathMap := map[string]string{} // key is origin, value is next hop to final destination
	tasks := list.New()
	tasks.PushBack(to)
	success := false
	for {
		front := tasks.Front()
		if front == nil {
			break
		}
		current := front.Value.(string)
		tasks.Remove(front)

		hopsToCurrent, ok := toFromMap[current]
		if !ok {
			continue
		}
		for _, hop := range hopsToCurrent {
			_, alreadyVisited := pathMap[hop]
			if alreadyVisited {
				continue
			}
			pathMap[hop] = current
			if hop == from {
				success = true
				break
			} else {
				tasks.PushBack(hop)
			}
		}
	}
	if !success {
		return nil, errors.New("No migration path found")
	}

	path := []*entity.Migration{}

	for current := from; current != to; current = pathMap[current] {
		path = append(path, migMap[pathMap[current]][current])
	}

	return path, nil
}
