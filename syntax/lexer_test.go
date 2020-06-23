package syntax

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLexerTags(t *testing.T) {
	items := lex("lexer", "#a1 #podróż")
	assert.Equal(t, item{
		typ: itemTag,
		val: "#a1",
	}, <-items)
	assert.Equal(t, item{
		typ: itemTag,
		val: "#podróż",
	}, <-items)
}

func TestLexerParensAndOr(t *testing.T) {
	items := lex("lexer", "(#mat|#klær)")
	expectedItems := []item{
		item{itemParenLeft, "("},
		item{itemTag, "#mat"},
		item{itemOr, "|"},
		item{itemTag, "#klær"},
		item{itemParenRight, ")"},
	}
	for _, expectedItem := range expectedItems {
		assert.Equal(t, expectedItem, <-items)
	}
}

func TestLexerPrefix(t *testing.T) {
	items := lex("lexer", "lang:no (#mat|#klær)")
	expectedItems := []item{
		item{itemOp, "lang:no"},
		item{itemParenLeft, "("},
		item{itemTag, "#mat"},
		item{itemOr, "|"},
		item{itemTag, "#klær"},
		item{itemParenRight, ")"},
	}
	for _, expectedItem := range expectedItems {
		assert.Equal(t, expectedItem, <-items)
	}
}
