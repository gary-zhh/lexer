package sql

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	name  string
	input string    // the string being scanned
	start int       // start position of this item
	pos   int       // current position in the input
	width int       // width of last rune read
	items chan item // channel of scanned items
}
type stateFunc func(*lexer) stateFunc

func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}
func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.items)
}
func (l *lexer) emit(t itemType) {
	l.items <- item{typ: t, val: l.input[l.start:l.pos]}
	l.start = l.pos
}
func (l *lexer) next() (r rune) {
	if l.pos > len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// backup only called after next is called
func (l *lexer) backup() {
	l.pos -= l.width
}
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}
func (l *lexer) ignore() {
	l.start = l.pos
}
func (l *lexer) errorf(format string, args ...interface{}) stateFunc {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}
func (l *lexer) skipSpace() bool {
	skipped := false
	for {
		if l.accept(Space) {
			skipped = true
			continue
		}
		break
	}
	l.ignore()
	return skipped
}
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}
func lexStart(l *lexer) stateFunc {
	l.skipSpace()
	if strings.HasPrefix(l.input[l.pos:], KeySelect) {
		l.pos += len(KeySelect)
		l.emit(itemSelect)
		if !l.skipSpace() {
			return l.errorf("syntax error: start with %q", l.input[l.pos])
		}
		return lexField
	}
	return l.errorf("syntax error: start with %q", l.input[l.pos])
}

func lexField(l *lexer) stateFunc {
	l.skipSpace()

	for {
		if r := l.next(); unicode.IsLetter(r) {
			l.acceptRun(letter)
			if l.accept(".") {
				if !unicode.IsLetter(l.peek()) {
					return l.errorf("syntax error: query field %q end with '.'", l.input[l.pos:])
				}
				l.acceptRun(letter)
			}
			l.emit(itemIdentifier)
			l.skipSpace()
			if !l.accept(MakrComma) {
				break
			} else {
				l.emit(itemComma)
				l.skipSpace()
			}
		} else {
			return l.errorf("syntax error: query field %q not valid", l.input[l.pos:])
		}
	}
	if strings.HasPrefix(l.input[l.pos:], KeyFrom) {
		l.pos += len(KeyFrom)
		if l.accept(Space) {
			l.backup()
			l.emit(itemFrom)
			return lexFrom
		} else {
			l.pos -= len(KeyFrom)
		}

	}
	if strings.HasPrefix(l.input[l.pos:], KeyWhere) {
		l.pos += len(KeyWhere)
		if l.accept(Space) {
			l.backup()
			l.emit(itemWhere)
			return lexCondition
		} else {
			l.pos -= len(KeyWhere)
		}

	}
}

func lexFrom(l *lexer) stateFunc {

}
