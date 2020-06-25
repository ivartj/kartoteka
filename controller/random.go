package controller

import (
	"database/sql"
	"github.com/ivartj/kartoteka/core"
	entity "github.com/ivartj/kartoteka/core/entity"
	"github.com/ivartj/kartoteka/repository"
	"github.com/ivartj/kartoteka/service"
	"github.com/ivartj/kartoteka/syntax"
	"html/template"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Random struct {
	txProvider
	templateProvider
	mux      *http.ServeMux
	rng      *rand.Rand
	rngMutex sync.Mutex
}

func NewRandom(db *sql.DB, tpl *template.Template) *Random {
	ctx := &Random{
		mux:              http.NewServeMux(),
		txProvider:       txProvider{db},
		templateProvider: templateProvider{tpl},
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
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
		err = ctx.Template().ExecuteTemplate(w, "random-index", nil)
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

		pageData := map[string]interface{}{
			"Spec": q,
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
			pageData["Error"] = "No word entries match that query"
			goto render
		} else if err != nil {
			panic(err)
		}
		pageData["Word"] = word

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
