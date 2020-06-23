package syntax

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type itemType int

const (
	eof = 0
)

const (
	itemError itemType = iota
	itemEOF
	itemTag        // #a1, #mat
	itemOp         // lang:no, tr:pl
	itemOr         // |, OR
	itemParenLeft  // (
	itemParenRight // )
)

type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type lexer struct {
	name  string
	input string
	items chan item
	start int
	pos   int
	width int
}

type stateFn func(*lexer) stateFn

func lex(name, input string) <-chan item {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l.items
}

func (l *lexer) run() {
	for state := startState; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	var r rune
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) accept(predicate func(rune) bool) bool {
	if predicate(l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRune(r rune) bool {
	if r == l.next() {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(predicate func(rune) bool) bool {
	if !predicate(l.next()) {
		l.backup()
		return false
	}
	for predicate(l.next()) {
	}
	l.backup()
	return true
}

func (l *lexer) errorf(format string, v ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, v...),
	}
	return nil
}

func startState(l *lexer) stateFn {
	l.acceptRun(func(r rune) bool { return unicode.IsSpace(r) })
	l.ignore()
	switch r := l.next(); {
	case r == '#':
		return tagState
	case r == '|':
		l.emit(itemOr)
		return startState
	case r == '(':
		l.emit(itemParenLeft)
		return startState
	case r == ')':
		l.emit(itemParenRight)
		return startState
	case unicode.IsLetter(r) && r <= unicode.MaxASCII:
		return opState
	case r == eof:
		return nil
	default:
		return l.errorf("Unhandled input case '%c'", r)
	}
}

func validTagRune(r rune) bool {
	return unicode.IsLetter(r) || r == '-' || unicode.IsNumber(r)
}

func isTermDelimiter(r rune) bool {
	return unicode.IsSpace(r) || r == eof || r == '|' || r == ')' || r == '('
}

func tagState(l *lexer) stateFn {
	switch r := l.next(); {
	case unicode.IsLetter(r):
		for r = l.next(); validTagRune(r); r = l.next() {
		}
		l.backup()
		if !(isTermDelimiter(r)) {
			return l.errorf("Unexpected symbol '%c' in tag", r)
		}
		l.emit(itemTag)
		return startState
	default:
		return l.errorf("A hashtag symbol (#) needs to be followed by a letter")
	}
}

func opState(l *lexer) stateFn {
	l.acceptRun(func(r rune) bool { return unicode.IsLetter(r) && r < unicode.MaxASCII })
	r := l.next()
	if r != ':' {
		return l.errorf("Expected colon (:), got '%c'", r)
	}
	// TODO: Allow for arguments enclosed in quotes
	l.acceptRun(func(r rune) bool { return unicode.IsLetter(r) })
	if r := l.peek(); !isTermDelimiter(r) {
		return l.errorf("Unexpected symbol '%c' in search operation", r)
	}
	op := l.input[l.start:l.pos]
	if _, err := opToSpec(op); err != nil {
		return l.errorf("%s is not valid prefix operator", op)
	}
	l.emit(itemOp)
	return startState
}
