package syntax

import (
	"fmt"
	"github.com/ivartj/kartotek/core"
	"strings"
)

func tagToSpec(tag string) core.TagWordSpec {
	return core.TagWordSpec(tag[1:])
}

func opToSpec(opString string) (core.WordSpec, error) {
	colon := strings.IndexRune(opString, ':')
	op := opString[:colon]
	arg := opString[colon+1:]
	switch op {
	case "lang":
		return core.LanguageWordSpec(arg), nil
	case "tr":
		return core.TranslationWordSpec(arg), nil
	default:
		return nil, fmt.Errorf("Unrecognized search operator '%s'", op)
	}
}
