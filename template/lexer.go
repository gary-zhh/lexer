package template

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type stateFunc func(*lexer) stateFunc

const (
	leftMeta  = "{{"
	rightMeta = "{{"
	eof       = 0
)

type lexer struct {
	name  string
	input string    // the string being scanned
	start int       // start position of this item
	pos   int       // current position in the input
	width int       // width of last rune read
	items chan item // channel of scanned items
}

func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l, l.items
}

func (l *lexer) emit(t itemType) {
	l.items <- item{typ: t, val: l.input[l.start:l.pos]}
	l.start = l.pos
}
func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func lexText(l *lexer) stateFunc {
	for {
		if strings.HasPrefix(l.input[l.pos:], leftMeta) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexLeftMeta
		}
		if l.next() == eof {
			break
		}
	}
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
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
func lexLeftMeta(l *lexer) stateFunc {
	l.pos += len(leftMeta)

	l.emit(itemLeftMeta)
	return lexInsideAction
}

func lexRightMeta(l *lexer) stateFunc {
	l.pos += len(rightMeta)
	l.emit(itemRightMeta)
	return lexText
}
func (l *lexer) ignore() {
	l.start = l.pos
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
func lexInsideAction(l *lexer) stateFunc {
	for {
		if strings.HasPrefix(l.input[l.pos:], rightMeta) {
			return lexRightMeta
		}
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("unclosed action")
		case unicode.IsSpace(r):
			l.ignore()
		case r == '|':
			l.emit(itemPipe)
		case r == '"':
			//return lexQuote
		case r == '`':
			//return lexRawQuote
		case r == '+' || r == '-' || '0' <= r && r <= '9':
			l.backup()
			return lexNumber
		case unicode.IsLetter(r):
			l.backup()
			//return lexIdentifier
		}
	}
}

func lexNumber(l *lexer) stateFunc {
	l.accept("+-")
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits += "abcdefABCDEF"
	}

	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}

	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}

	l.accept("i")
	if unicode.IsLetter(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(itemNumber)
	return lexInsideAction
}
func (l *lexer) errorf(format string, args ...interface{}) stateFunc {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}
