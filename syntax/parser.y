%{
	package syntax

	import (
		"github.com/ivartj/kartotek/core"
		"fmt"
		"strings"
		"errors"
	)

%}

%union {
	spec core.WordSpec
	token item
}

%type <spec> top expr expr1 expr2
%token <token> TOKEN_TAG TOKEN_OP 
%token TOKEN_ERROR
%token '|' '(' ')'

%%
	top:
		expr
		{
			yylex.(*parser).val = $1
		}

	expr:
		expr1
	| expr '|' expr1
		{
			$$ = &core.OrWordSpec{
				Left: $1,
				Right: $3,
			}
		}

	expr1:
		expr2
	| expr1 expr2
		{
			$$ = &core.AndWordSpec{
				Left: $1,
				Right: $2,
			}
		}
	
	expr2:
		TOKEN_TAG
		{
			$$ = tagToSpec($1.val)
		}
	| TOKEN_OP
		{
			spec, _ := opToSpec($1.val)
			$$ = spec
		}
	| '(' expr ')'
		{
			$$ = $2
		}
%%

type parser struct {
	items <-chan item
	val core.WordSpec
	errlog strings.Builder
}

func (l *parser) Lex(lval *yySymType) int {
	i, ok := <-l.items
	if !ok {
		return eof
	}
	switch i.typ {
	case itemParenLeft:
		return '('
	case itemParenRight:
		return ')'
	case itemOr:
		return '|'
	case itemTag:
		lval.token = i
		return TOKEN_TAG
	case itemOp:
		lval.token = i
		return TOKEN_OP
	case itemError:
		return TOKEN_ERROR
	default:
		panic(fmt.Errorf("Unrecognized token type %d", i.typ))
	}
}

func (l *parser) Error(s string) {
	l.errlog.WriteString(s)
	l.errlog.WriteRune('\n')
}

func ParseWordSpec(input string) (core.WordSpec, error) {
	l := &parser{items: lex("lexer", input)}
	if yyParse(l) != 0 {
		return nil, errors.New(l.errlog.String())
	}
	return l.val, nil
}

