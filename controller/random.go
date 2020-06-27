package controller

import (
	"database/sql"
	"fmt"
	"github.com/ivartj/kartoteka/core"
	entity "github.com/ivartj/kartoteka/core/entity"
	"github.com/ivartj/kartoteka/repository"
	"github.com/ivartj/kartoteka/service"
	"github.com/ivartj/kartoteka/syntax"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"html/template"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Random struct {
	txProvider
	templateProvider
	mux        *http.ServeMux
	rng        *rand.Rand
	rngMutex   sync.Mutex
	i18nBundle *i18n.Bundle
}

func NewRandom(db *sql.DB, tpl *template.Template, i18nBundle *i18n.Bundle) *Random {
	ctx := &Random{
		mux:              http.NewServeMux(),
		txProvider:       txProvider{db},
		templateProvider: templateProvider{tpl},
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
		i18nBundle:       i18nBundle,
	}

	return ctx
}

type txProvider struct {
	*sql.DB
}

func (self txProvider) Tx() *sql.Tx {
	tx, err := self.Begin()
	if err != nil {
		panic(err)
	}
	return tx
}

type templateProvider struct {
	tpl *template.Template
}

func (self templateProvider) Template() *template.Template {
	return self.tpl
}

func (ctx *Random) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query().Get("q")
	var err error
	if q == "" {
		localizer := i18n.NewLocalizer(ctx.i18nBundle, req.Header.Get("Accept-Language"))
		pageData := map[string]interface{}{
			"Localizer": localizer,
		}
		err = ctx.Template().ExecuteTemplate(w, "random-index", pageData)
		if err != nil {
			panic(err)
		}
	} else {
		var wordStore core.WordStore
		var wordLottery core.WordLottery
		var word *entity.Word
		var wordSpec core.WordSpec
		var tx *sql.Tx
		var languageNativeNameMap map[string]string

		localizer := i18n.NewLocalizer(ctx.i18nBundle, req.Header.Get("Accept-Language"))

		pageData := map[string]interface{}{
			"Spec":      q,
			"Localizer": localizer,
		}

		wordSpec, err = syntax.ParseWordSpec(q)
		if err != nil {
			pageData["Error"] = err.Error()
			goto render
		}

		tx = ctx.Tx()
		defer tx.Rollback()

		languageNativeNameMap, err = service.NewLanguageService(repository.NewLanguageStore(tx)).GetNativeNameMap()
		if err != nil {
			panic(err)
		}
		pageData["LanguageNativeNameMap"] = languageNativeNameMap

		wordStore = repository.NewWordStore(tx)
		wordLottery = service.NewWordLottery(wordStore, wordSpec, ctx.rng)
		word, err = wordLottery.DrawWord()
		if err == core.ErrNotFound {
			msg, err := localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "NoMatches",
					Other: "No entries match that query",
				},
			})
			if err != nil {
				panic(fmt.Errorf("Localization error: %w", err))
			}
			pageData["Error"] = msg
			goto render
		} else if err != nil {
			panic(err)
		}
		pageData["Word"] = word
		pageData["Localizer"] = i18n.NewLocalizer(ctx.i18nBundle, word.LanguageCode, req.Header.Get("Accept-Language"))

	render:
		err = ctx.Template().ExecuteTemplate(w, "random-word", pageData)
		if err != nil {
			panic(err)
		}
		if tx != nil {
			tx.Commit()
		}
	}
}
