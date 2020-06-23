package service

import (
	"fmt"
	"github.com/ivartj/kartotek/core"
	entity "github.com/ivartj/kartotek/core/entity"
)

type WordLottery struct {
	wordStore core.WordStore
	rng       core.Rand
	spec      core.WordSpec
}

func NewWordLottery(wordStore core.WordStore, spec core.WordSpec, rng core.Rand) *WordLottery {
	return &WordLottery{
		wordStore: wordStore,
		spec:      spec,
		rng:       rng,
	}
}

func (lot *WordLottery) DrawWord() (*entity.Word, error) {
	query := core.WordQuery{
		Spec: lot.spec,
	}
	numberOfCards, err := lot.wordStore.Count(&query)
	if err != nil {
		return nil, fmt.Errorf("Error getting a count of matching words: %w", err)
	}
	if numberOfCards == 0 {
		return nil, core.ErrNotFound
	}
	drawnCardNumber := lot.rng.Int() % numberOfCards
	query.SetRange(drawnCardNumber, 1)
	cards, err := lot.wordStore.List(&query)
	if err != nil {
		return nil, fmt.Errorf("Error getting a word at a random offset: %w", err)
	}
	return cards[0], nil
}
