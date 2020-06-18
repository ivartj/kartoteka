package repository

import (
	"database/sql"
	"fmt"
	"github.com/ivartj/kartotek/core"
	entity "github.com/ivartj/kartotek/core/entity"
	util "github.com/ivartj/kartotek/util"
	sqlutil "github.com/ivartj/kartotek/util/sqlutil"
	"strings"
)

type WordStore struct {
	db core.DB
}

func NewWordStore(db core.DB) *WordStore {
	return &WordStore{
		db: db,
	}
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (repo *WordStore) Get(id entity.ID) (*entity.Word, error) {
	stmtSql := `
		select id, word, language_code, image_id, notes from word where id = :id;`

	var word entity.Word

	row := repo.db.QueryRow(stmtSql, sql.Named("id", id))
	err := sqlutil.Row{row}.ScanEntity(&word)
	if err == sql.ErrNoRows {
		return nil, core.ErrNotFound
	}
	if err != nil {
	}

	return &word, nil
}

func (repo *WordStore) Add(word *entity.Word) error {
	err := sqlutil.DB{repo.db}.InsertEntity("word", word)
	if err != nil {
		return err
	}
	if word.Tags != nil && len(word.Tags) != 0 {
		b := new(util.FormatBuilder)
		b.Add("INSERT INTO word_tag (word_id, tag) VALUES")
		for i, tag := range word.Tags {
			if i != 0 {
				b.Add(",")
			}
			b.Add(" (?, ?)", word.ID, tag)
		}
		_, err = repo.db.Exec(b.Format(), b.Args()...)
		if err != nil {
			return err
		}
	}

	if word.Translations != nil && len(word.Translations) != 0 {
		var b util.FormatBuilder
		b.Add("INSERT INTO word_translation (word_id, language_code, translation) VALUES")
		for i, tr := range word.Translations {
			if i != 0 {
				b.Add(",")
			}
			b.Add(" (?, ?, ?)", word.ID, tr.LanguageCode, tr.Translation)
			_, err = repo.db.Exec(b.Format(), b.Args()...)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (repo *WordStore) Update(word *entity.Word) error {
	err := sqlutil.DB{repo.db}.InsertOrReplaceEntity("word", word)
	if err != nil {
		return err
	}
	return err
}

func (repo *WordStore) Delete(id entity.ID) error {
	_, err := repo.db.Exec("delete from word where word_id = ?;", id)
	return err
}

func (repo *WordStore) List(query *core.WordQuery) ([]*entity.Word, error) {
	querySql, args := wordQuerySql(query, "*")
	fmt.Println(querySql)
	rows, err := repo.db.Query(querySql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	words := []*entity.Word{}
	for rows.Next() {
		word := new(entity.Word)
		err = sqlutil.Rows{rows}.ScanEntity("", word)
		if err != nil {
			return nil, err
		}
		words = append(words, word)
	}
	return words, nil
}

func (repo *WordStore) Count(query *core.WordQuery) (int, error) {
	querySql, args := wordQuerySql(query, "word.*")
	row := repo.db.QueryRow(querySql, args...)
	var count int
	err := row.Scan(&count)
	return count, err
}

func escapeSqlLikeString(s string, escape rune) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case escape:
			sb.WriteRune(escape)
			sb.WriteRune(escape)
		case '%':
			sb.WriteRune(escape)
			sb.WriteRune('%')
		case '_':
			sb.WriteRune(escape)
			sb.WriteRune('_')
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func wordQuerySql(query *core.WordQuery, projection string) (string, []interface{}) {
	var b util.FormatBuilder

	b.Add("SELECT \n").Add(projection).Add(" FROM word_view\n")
	b.Add(" WHERE \n")

	wordQuerySqlWhereClause(&b, query.Spec)

	if query.HasRange() {
		b.Add(" LIMIT ? OFFSET ? \n", query.Length, query.Offset)
	}

	return b.Format(), b.Args()
}

func wordQuerySqlWhereClause(b *util.FormatBuilder, spec core.WordSpec) {
	b.Add(" (")
	switch s := spec.(type) {
	case core.AnyWordSpec:
		b.Add(" true")
	case *core.AndWordSpec:
		b.Add(" (")
		wordQuerySqlWhereClause(b, s.Left)
		b.Add(" and")
		wordQuerySqlWhereClause(b, s.Right)
		b.Add(" )")
	case *core.OrWordSpec:
		b.Add(" (")
		wordQuerySqlWhereClause(b, s.Left)
		b.Add(" or")
		wordQuerySqlWhereClause(b, s.Right)
		b.Add(" )")
	case *core.NotWordSpec:
		b.Add(" not (")
		wordQuerySqlWhereClause(b, s.Spec)
		b.Add(" )")
	case core.TagWordSpec:
		likePattern := fmt.Sprintf("%% %s %%", escapeSqlLikeString(string(s), '\\'))
		b.Add(" (' ' || tags || ' ') LIKE ? ESCAPE '\\'", likePattern)
	case core.TranslationWordSpec:
		likePattern := fmt.Sprintf("%% %s %%", escapeSqlLikeString(string(s), '\\'))
		b.Add(" (' ' || translation_codes || ' ') LIKE ? ESCAPE '\\'", likePattern)
	case core.LanguageWordSpec:
		b.Add(" language_code is ?", string(s))
	case core.UserWordSpec:
		b.Add(" username is ?", string(s))
	}
	b.Add(" )")
	fmt.Println(b.Format())
}
