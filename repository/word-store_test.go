package repository

import (
	"database/sql"
	"github.com/ivartj/kartotek/core"
	entity "github.com/ivartj/kartotek/core/entity"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWordStore(t *testing.T) {

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("pragma foreign_keys = on;")
	if err != nil {
		panic(err)
	}

	err = InitSchema(db)
	if err != nil {
		panic(err)
	}

	userStore := NewUserStore(db)
	languageStore := NewLanguageStore(db)
	wordStore := NewWordStore(db)

	for _, language := range []entity.Language{
		entity.Language{
			Code:       "no",
			NativeName: "norsk",
		},
		entity.Language{
			Code:       "pl",
			NativeName: "polski",
		},
	} {
		err = languageStore.Update(&language)
		if err != nil {
			panic(err)
		}
	}

	user := &entity.User{
		ID:              entity.UserID(entity.NewID()),
		Username:        "bob",
		Email:           "bob@example.com",
		EmailUnverified: "bob@example.com",
		PasswordHash:    "",
	}
	assert.NotNil(t, user.ID)
	err = userStore.Update(user)
	if err != nil {
		panic(err)
	}

	words := []*entity.Word{
		&entity.Word{
			ID:           entity.WordID(entity.NewID()),
			Word:         "et eple",
			LanguageCode: "no",
			UserID:       user.ID,
		},
		&entity.Word{
			ID:           entity.WordID(entity.NewID()),
			Word:         "jabłko",
			LanguageCode: "pl",
			UserID:       user.ID,
		},
	}
	for _, word := range words {
		err = wordStore.Update(word)
		if err != nil {
			t.Fatalf("Failed to add a word: %s", err)
		}
	}

	retWords, err := wordStore.List(new(core.WordQuery).SetLanguage("no"))
	if err != nil {
		t.Fatalf("Word query failed: %s", err)
	}

	assert.Equal(t, 1, len(retWords))
	assert.Equal(t, "et eple", retWords[0].Word)
	assert.Equal(t, words[0].ID, retWords[0].ID)
	assert.Equal(t, words[0].ImageID, retWords[0].ImageID)
	assert.Equal(t, words[0].UserID, retWords[0].UserID)
}
