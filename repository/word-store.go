package repository

import (
	"database/sql"
	"github.com/ivartj/kartotek/core"
	entity "github.com/ivartj/kartotek/core/entity"
	util "github.com/ivartj/kartotek/util"
	sqlutil "github.com/ivartj/kartotek/util/sqlutil"
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

func (repo *WordStore) Update(word *entity.Word) error {
	return sqlutil.DB{repo.db}.InsertOrReplaceEntity("word", word)
}

func (repo *WordStore) Delete(id entity.ID) error {
	_, err := repo.db.Exec("delete from word where word_id = ?;", id)
	return err
}

func (repo *WordStore) List(query *core.WordQuery) ([]*entity.Word, error) {
	querySql, args := wordQuerySql(query, "*")
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

func wordQuerySql(query *core.WordQuery, projection string) (string, []interface{}) {
	var b util.FormatBuilder

	if query.HasTags() {
		b.Add("with required_tags as (values ")
		for i, tag := range query.Tags {
			if i != 0 {
				b.Add(", ")
			}
			b.Add("(?)", tag)
		}
		b.Add(") ")
	}

	b.Add("select ").Add(projection).Add(" from word ")

	if query.HasTags() {
		b.Add("natural join word_tag ")
	}

	b.Add("where true ")
	if query.HasLanguage() {
		b.Add("and language_code = ? ", query.LangCode)
	}
	if query.HasTags() {
		b.Add("word_tag.tag in required_tags ")
	}

	if query.HasTags() {
		b.Add("group by word_id having count(*) = (select count(*) from required_tags) ")
	}

	if query.HasRange() {
		b.Add("limit ? offset ? ", query.Length, query.Offset)
	}

	return b.Format(), b.Args()
}
