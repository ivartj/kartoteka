package repository

import (
	"github.com/ivartj/kartoteka/core"
	entity "github.com/ivartj/kartoteka/core/entity"
	"github.com/ivartj/kartoteka/util/sqlutil"
)

type LanguageStore struct {
	db core.DB
}

func NewLanguageStore(db core.DB) *LanguageStore {
	return &LanguageStore{
		db: db,
	}
}

func (store *LanguageStore) Get(langCode string) (*entity.Language, error) {
	row := store.db.QueryRow("select * from language where lang_code = ?;", langCode)
	var language entity.Language
	err := sqlutil.Row{row}.ScanEntity(&language)
	if err != nil {
		return nil, err
	}
	return &language, nil
}

func (store *LanguageStore) ListAll() ([]*entity.Language, error) {
	rows, err := store.db.Query("select * from language;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	languages := []*entity.Language{}
	for rows.Next() {
		language := new(entity.Language)
		err = sqlutil.Rows{rows}.ScanEntity("", language)
		if err != nil {
			return nil, err
		}
		languages = append(languages, language)
	}
	return languages, nil
}

func (store *LanguageStore) Update(language *entity.Language) error {
	return sqlutil.DB{store.db}.InsertOrReplaceEntity("language", language)
}

func (store *LanguageStore) Delete(langCode string) error {
	_, err := store.db.Exec("delete from language where lang_code = ?;", langCode)
	return err
}
