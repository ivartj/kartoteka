package entities

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshalWordTranslation(t *testing.T) {
	jsonString := `
		[
			{
				"word_id": "23456789",
				"language_code": "en",
				"translation": "an apple"
			},
			{
				"word_id": "23456789",
				"language_code": "no",
				"translation": "et eple"
			}
		]`
	var translations []*WordTranslation

	err := json.Unmarshal([]byte(jsonString), &translations)
	if err != nil {
		t.Fatalf("Failed to unmarshall JSON translations: %s", err)
	}

	assert.Equal(t, 2, len(translations))
	assert.Equal(t, "en", translations[0].LanguageCode)
	assert.Equal(t, "no", translations[1].LanguageCode)
}
