package core

import (
	entity "github.com/ivartj/kartotek/core/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWordSpecBasic(t *testing.T) {
	words := []*entity.Word{
		&entity.Word{
			LanguageCode: "no",
			Tags:         []string{"a1", "mat"},
			UserUsername: "ivartj",
		},
		&entity.Word{
			LanguageCode: "nb",
			Tags:         []string{"a1", "mat"},
			UserUsername: "ivartj",
		},
	}

	spec := &AndWordSpec{
		Left: TagWordSpec("a1"),
		Right: &AndWordSpec{
			Left:  UserWordSpec("ivartj"),
			Right: LanguageWordSpec("no"),
		},
	}

	assert.True(t, spec.Match(words[0]))
	assert.False(t, spec.Match(words[1]))
}
