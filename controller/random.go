package controller

import (
	"database/sql"
	"github.com/ivartj/kartotek/repository"
	"github.com/ivartj/kartotek/service"
	"github.com/ivartj/kartotek/syntax"
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
		tx := ctx.Tx()
		defer tx.Rollback()
		wordStore := repository.NewWordStore(tx)
		wordSpec, err := syntax.ParseWordSpec(q)
		if err != nil {
			// TODO: report to user
			panic(err)
		}
		wordLottery := service.NewWordLottery(wordStore, wordSpec, ctx.rng)
		word, err := wordLottery.DrawWord()
		if err != nil {
			// TODO: report to user in case of core.ErrNotFound
			panic(err)
		}
		err = ctx.Template().ExecuteTemplate(w, "random-word", map[string]interface{}{
			"Word": word,
			"Spec": q,
		})
		if err != nil {
			panic(err)
		}
		tx.Commit()
	}
}
