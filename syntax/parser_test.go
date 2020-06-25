package syntax

import (
	"github.com/ivartj/kartoteka/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse1(t *testing.T) {
	input := "lang:no (#a1|#a2)"
	spec, err := ParseWordSpec(input)
	if err != nil {
		t.Fatalf("Parse failed: %s", err)
	}
	assert.Equal(t, &core.AndWordSpec{
		Left: core.LanguageWordSpec("no"),
		Right: &core.OrWordSpec{
			Left:  core.TagWordSpec("a1"),
			Right: core.TagWordSpec("a2"),
		},
	}, spec)
}

func TestParse2(t *testing.T) {
	input := "lang:no #a1 | lang:no #a2"
	spec, err := ParseWordSpec(input)
	if err != nil {
		t.Fatalf("Parse failed: %s", err)
	}
	assert.Equal(t, &core.OrWordSpec{
		Left: &core.AndWordSpec{
			Left:  core.LanguageWordSpec("no"),
			Right: core.TagWordSpec("a1"),
		},
		Right: &core.AndWordSpec{
			Left:  core.LanguageWordSpec("no"),
			Right: core.TagWordSpec("a2"),
		},
	}, spec)
}
