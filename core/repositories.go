package core

import (
	entity "github.com/ivartj/kartotek/core/entity"
)

type UserStore interface {
	Get(id entity.ID) (*entity.User, error)
	Update(word *entity.User) error
	Delete(id entity.ID) error
}

type WordStore interface {
	Get(id entity.ID) (*entity.Word, error)
	Add(word *entity.Word) error
	Update(word *entity.Word) error
	Delete(id entity.ID) error
	List(query *WordQuery) ([]*entity.Word, error)
	Count(query *WordQuery) (int, error)
}

type WordQuery struct {
	Spec     WordSpec
	hasRange bool
	Offset   int
	Length   int
}

func (q *WordQuery) SetRange(offset, length int) *WordQuery {
	q.Offset = offset
	q.Length = length
	q.hasRange = true
	return q
}

func (q *WordQuery) HasRange() bool {
	return q.hasRange
}

type LanguageStore interface {
	Get(langCode string) (*entity.Language, error)
	ListAll() ([]*entity.Language, error)
	Update(language *entity.Language) error
	Delete(langCode string) error
}
