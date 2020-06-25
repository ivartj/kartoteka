package core

import (
	entity "github.com/ivartj/kartoteka/core/entity"
)

type WordLottery interface {
	DrawWord() (*entity.Word, error)
}

type Rand interface {
	Int() int
}

type LanguageService interface {
	GetNativeNameMap() (map[string]string, error)
}
