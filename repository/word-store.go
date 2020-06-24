package repository

import (
	"database/sql"
	"encoding/json"
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

func (repo *WordStore) Add(word *entity.Word) error {
	err := sqlutil.DB{repo.db}.InsertEntity("word", word)
	if err != nil {
		return err
	}

	err = repo.addTagsAndTranslations(word)
	if err != nil {
		return err
	}

	return nil
}

func (repo *WordStore) addTagsAndTranslations(word *entity.Word) error {
	var err error
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
		}
		_, err = repo.db.Exec(b.Format(), b.Args()...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *WordStore) deleteTagsAndTranslations(word *entity.Word) error {
	_, err := repo.db.Exec("DELETE FROM word_tag WHERE word_id = ?", word.ID)
	if err != nil {
		return err
	}
	_, err = repo.db.Exec("DELETE FROM word_translation WHERE word_id = ?", word.ID)
	if err != nil {
		return err
	}
	return nil
}

func (repo *WordStore) Update(word *entity.Word) error {
	err := sqlutil.DB{repo.db}.InsertOrReplaceEntity("word", word)
	if err != nil {
		return err
	}
	err = repo.deleteTagsAndTranslations(word)
	if err != nil {
		return err
	}
	err = repo.addTagsAndTranslations(word)
	if err != nil {
		return err
	}
	return nil
}

func (repo *WordStore) Delete(id entity.WordID) error {
	_, err := repo.db.Exec("delete from word where word_id = ?;", id)
	return err
}

func (repo *WordStore) Get(id entity.WordID) (*entity.Word, error) {
	rows, err := repo.db.Query("SELECT * FROM word_view WHERE word_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ok := rows.Next()
	if !ok {
		return nil, core.ErrNotFound
	}
	var word entity.Word
	err = scanWord(rows, "", &word)
	if err != nil {
		return nil, err
	}
	return &word, nil
}

func scanWord(rows *sql.Rows, columnPrefix string, word *entity.Word) error {
	rowMap := map[string]interface{}{}
	err := sqlutil.Rows{rows}.ScanMap(rowMap)
	if err != nil {
		return err
	}
	err = sqlutil.ScanEntityFromMap(rowMap, columnPrefix, word)
	if err != nil {
		return err
	}
	translationsJSON, ok := rowMap["translations"].(string)
	if !ok {
		return fmt.Errorf("Failed to cast %s to string", rowMap["translations"])
	}
	err = json.Unmarshal([]byte(translationsJSON), &word.Translations)
	if err != nil {
		return err
	}
	tagString, ok := rowMap["tags"].(string)
	if !ok {
		return fmt.Errorf("Failed to cast %s to string", rowMap["tags"])
	}
	word.Tags = strings.Split(tagString, " ")
	return nil
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
		err = scanWord(rows, "", word)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan word from database: %w", err)
		}
		words = append(words, word)
	}
	return words, nil
}

func (repo *WordStore) Count(query *core.WordQuery) (int, error) {
	querySql, args := wordQuerySql(query, "count(*)")
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
}
