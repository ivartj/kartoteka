package service

import (
	"github.com/ivartj/kartoteka/core"
)

type LanguageService struct {
	languageStore core.LanguageStore
}

func NewLanguageService(languageStore core.LanguageStore) *LanguageService {
	return &LanguageService{
		languageStore: languageStore,
	}
}

func (service *LanguageService) GetNativeNameMap() (map[string]string, error) {
	languages, err := service.languageStore.ListAll()
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(languages))
	for _, language := range languages {
		m[language.Code] = language.NativeName
	}
	return m, nil
}
