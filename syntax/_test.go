package syntax

import (
	"github.com/ivartj/kartotek/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	spec, err := Parse("lang:no (#a1 | #a2 tr:en)")
	if err != nil {
		t.Fatalf("Failed to parse search string: %s", err)
	}
	assert.Equal(t, &core.AndWordSpec{
		Left: core.LanguageWordSpec("no"),
		Right: &core.OrWordSpec{
			Left: core.TagWordSpec("a1"),
			Right: &core.AndWordSpec{
				Left:  core.TagWordSpec("a1"),
				Right: core.TranslationWordSpec("en"),
			},
		},
	}, spec)
}
