package repository

import (
	"database/sql"
	"github.com/ivartj/kartoteka/core"
	entity "github.com/ivartj/kartoteka/core/entity"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testContext struct {
	db        *sql.DB
	wordStore core.WordStore
	bobID     entity.UserID
	aliceID   entity.UserID
}

func newTestContext() *testContext {
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
			NativeName: "Norsk",
		},
		entity.Language{
			Code:       "nn",
			NativeName: "Nynorsk",
		},
		entity.Language{
			Code:       "pl",
			NativeName: "Polski",
		},
		entity.Language{
			Code:       "en",
			NativeName: "English",
		},
	} {
		err = languageStore.Update(&language)
		if err != nil {
			panic(err)
		}
	}

	users := []*entity.User{
		&entity.User{
			ID:              entity.UserID(entity.NewID()),
			Username:        "bob",
			Email:           "bob@example.com",
			EmailUnverified: "bob@example.com",
			PasswordHash:    "",
		},
		&entity.User{
			ID:              entity.UserID(entity.NewID()),
			Username:        "alice",
			Email:           "alice@example.com",
			EmailUnverified: "alice@example.com",
			PasswordHash:    "",
		},
	}
	for _, user := range users {
		err = userStore.Update(user)
		if err != nil {
			panic(err)
		}
	}

	return &testContext{
		db:        db,
		wordStore: wordStore,
		bobID:     users[0].ID,
		aliceID:   users[1].ID,
	}
}

func TestWordStoreBasic(t *testing.T) {
	ctx := newTestContext()
	defer ctx.db.Close()
	wordStore := ctx.wordStore
	bobID := ctx.bobID
	aliceID := ctx.aliceID

	words := []*entity.Word{
		&entity.Word{
			ID:           entity.WordID(entity.NewID()),
			Word:         "et eple",
			LanguageCode: "no",
			UserID:       bobID,
		},
		&entity.Word{
			ID:           entity.WordID(entity.NewID()),
			Word:         "jabłko",
			LanguageCode: "pl",
			UserID:       aliceID,
		},
	}
	for _, word := range words {
		err := wordStore.Add(word)
		if err != nil {
			t.Fatalf("Failed to add a word: %s", err)
		}
	}

	retWords, err := wordStore.List(&core.WordQuery{Spec: core.LanguageWordSpec("no")})
	if err != nil {
		t.Fatalf("Word query failed: %s", err)
	}

	assert.Equal(t, 1, len(retWords))
	assert.Equal(t, "et eple", retWords[0].Word)
	assert.Equal(t, words[0].ID, retWords[0].ID)
	assert.Equal(t, words[0].ImageID, retWords[0].ImageID)
	assert.Equal(t, words[0].UserID, retWords[0].UserID)
}

func TestWordStoreQueryLogicalOperators(t *testing.T) {
	ctx := newTestContext()
	defer ctx.db.Close()
	wordStore := ctx.wordStore
	bobID := ctx.bobID
	aliceID := ctx.aliceID

	words := []*entity.Word{
		&entity.Word{
			ID:           entity.WordID(entity.NewID()),
			Word:         "et eple",
			LanguageCode: "no",
			UserID:       bobID,
			Tags:         []string{"a1", "mat"},
			Translations: []*entity.WordTranslation{
				&entity.WordTranslation{
					LanguageCode: "pl",
					Translation:  "jabłko",
				},
			},
		},
		&entity.Word{
			ID:           entity.WordID(entity.NewID()),
			Word:         "en gulrot",
			LanguageCode: "no",
			UserID:       aliceID,
			Tags:         []string{"a1", "mat"},
		},
	}
	for _, word := range words {
		err := wordStore.Add(word)
		if err != nil {
			t.Fatalf("Failed to add a word: %s", err)
		}
	}

	retWords, err := wordStore.List(&core.WordQuery{
		Spec: &core.AndWordSpec{
			Left:  core.LanguageWordSpec("no"),
			Right: core.TranslationWordSpec("pl"),
		},
	})
	if err != nil {
		t.Fatalf("Word query failed: %s", err)
	}

	assert.Equal(t, 1, len(retWords))
	assert.Equal(t, "et eple", retWords[0].Word)

	retWords, err = wordStore.List(&core.WordQuery{
		Spec: &core.OrWordSpec{
			Left:  core.UserWordSpec("alice"),
			Right: core.TranslationWordSpec("pl"),
		},
	})
	if err != nil {
		t.Fatalf("Word query failed: %s", err)
	}

	assert.Equal(t, 2, len(retWords))
	assert.NotEqual(t, retWords[0].ID, retWords[1].ID)
}

func TestWordStoreListWithTranslations(t *testing.T) {
	ctx := newTestContext()
	defer ctx.db.Close()
	wordStore := ctx.wordStore
	bobID := ctx.bobID
	aliceID := ctx.aliceID

	words := []*entity.Word{
		&entity.Word{
			ID:           entity.WordID(entity.NewID()),
			Word:         "et eple",
			LanguageCode: "no",
			UserID:       bobID,
			Tags:         []string{"a1", "mat"},
			Translations: []*entity.WordTranslation{
				&entity.WordTranslation{
					LanguageCode: "pl",
					Translation:  "jabłko",
				},
				&entity.WordTranslation{
					LanguageCode: "en",
					Translation:  "an apple",
				},
			},
		},
		&entity.Word{
			ID:           entity.WordID(entity.NewID()),
			Word:         "en gulrot",
			LanguageCode: "no",
			UserID:       aliceID,
			Tags:         []string{"a1", "mat"},
		},
	}
	for _, word := range words {
		err := wordStore.Add(word)
		if err != nil {
			t.Fatalf("Failed to add a word: %s", err)
		}
	}

	retWords, err := wordStore.List(&core.WordQuery{Spec: core.AnyWordSpec{}})
	if err != nil {
		t.Fatalf("Failed to list words: %s", err)
	}

	assert.Equal(t, 2, len(retWords))
	var bobsWord *entity.Word
	for _, word := range retWords {
		if word.UserID == bobID {
			bobsWord = word
			break
		}
	}
	if bobsWord == nil {
		t.Fatalf("Failed to find the word entry by Bob")
	}
	assert.Equal(t, 2, len(bobsWord.Translations))
	for _, languageCode := range []string{"en", "pl"} {
		languageFound := false
		for _, translation := range bobsWord.Translations {
			if translation.LanguageCode == languageCode {
				languageFound = true
			} else {
				continue
			}
			switch languageCode {
			case "en":
				assert.Equal(t, "an apple", translation.Translation)
			case "pl":
				assert.Equal(t, "jabłko", translation.Translation)
			}
		}
		if !languageFound {
			t.Errorf("Did not find the expected language '%s' among the translations", languageCode)
		}
	}
}

func TestWordStoreUpdate(t *testing.T) {
	ctx := newTestContext()
	defer ctx.db.Close()
	wordStore := ctx.wordStore
	bobID := ctx.bobID
	//aliceID := ctx.aliceID

	word := &entity.Word{
		ID:           entity.WordID(entity.NewID()),
		Word:         "et eple",
		LanguageCode: "no",
		UserID:       bobID,
		Translations: []*entity.WordTranslation{
			&entity.WordTranslation{
				LanguageCode: "pl",
				Translation:  "jabłko",
			},
			&entity.WordTranslation{
				LanguageCode: "en",
				Translation:  "apple",
			},
		},
	}
	err := wordStore.Add(word)
	if err != nil {
		panic(err)
	}

	word.Translations[1].Translation = "an apple"
	word.Translations = append(word.Translations, &entity.WordTranslation{
		LanguageCode: "nn",
		Translation:  "eit eple",
	})
	err = wordStore.Update(word)
	if err != nil {
		t.Fatalf("Failed to update word: %s", err)
	}

	word, err = wordStore.Get(word.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated translation: %s", err)
	}

	assert.Equal(t, 3, len(word.Translations))
}
