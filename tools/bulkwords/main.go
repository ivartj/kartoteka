package main

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ivartj/kartotek/core"
	entity "github.com/ivartj/kartotek/core/entity"
	"github.com/ivartj/kartotek/repository"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

type word struct {
	Word  string            `toml:"word"`
	Lang  string            `toml:"lang"`
	Tr    map[string]string `toml:"tr"`
	Notes string            `toml:"notes"`
	Tags  []string          `toml:"tags"`
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "Usage: bulkwords <word-file> <database>")
}

func main() {
	if len(os.Args) != 3 {
		usage(os.Stderr)
		os.Exit(1)
	}
	wordFilename := os.Args[1]
	username := "bulkwords"
	databaseFilename := os.Args[2]

	db, err := sql.Open("sqlite3", databaseFilename)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer func() {
		err := recover()
		if err != nil {
			log.Fatal(err)
			tx.Rollback()
		}
	}()
	wordStore := repository.NewWordStore(tx)
	userStore := repository.NewUserStore(tx)
	user, err := userStore.GetByUsername(username)
	if err == core.ErrNotFound {
		user = &entity.User{
			ID:       entity.UserID(entity.NewID()),
			Username: "bulkwords",
		}
		userStore.Update(user)
	} else if err != nil {
		panic(err)
	}

	wordFileContent, err := ioutil.ReadFile(wordFilename)
	if err != nil {
		panic(err)
	}

	separator := regexp.MustCompile(`\n--\r?\n?`)
	wordSections := separator.Split(string(wordFileContent), -1)
	for _, wordSection := range wordSections {
		var w word
		err = toml.Unmarshal([]byte(wordSection), &w)
		if err != nil {
			panic(err)
		}
		word := &entity.Word{
			ID:           entity.WordID(entity.NewID()),
			Word:         w.Word,
			UserID:       user.ID,
			LanguageCode: w.Lang,
			Translations: []*entity.WordTranslation{},
			Tags:         w.Tags,
		}
		for key, value := range w.Tr {
			word.Translations = append(word.Translations, &entity.WordTranslation{
				LanguageCode: key,
				Translation:  value,
			})
		}
		err = wordStore.Add(word)
		if err != nil {
			panic(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}
