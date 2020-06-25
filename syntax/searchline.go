package syntax

import (
	"github.com/ivartj/kartoteka/core"
)

func Parse(line string) (core.WordSpec, error) {
	return &core.AndWordSpec{
		Left: core.LanguageWordSpec("no"),
		Right: &core.OrWordSpec{
			Left: core.TagWordSpec("a1"),
			Right: &core.AndWordSpec{
				Left:  core.TagWordSpec("a1"),
				Right: core.TranslationWordSpec("en"),
			},
		},
	}, nil
}
