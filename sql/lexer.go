package sql

import (
	"fmt"
	"strings"
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

func (l *lexer) nextTerm() string {
	l.skipSpace()
	l.acceptRun(letter)
	l.width = l.pos - l.start
	return l.input[l.start:l.pos]
}
func (l *lexer) nextTermWithDot() (string, bool) {
	l.skipSpace()
	l.acceptRun(letter)
	for l.accept(MarkDot) {
		if sub := l.nextTerm(); sub == "" {
			return l.input[l.start:l.pos], false
		}
	}
	l.width = l.pos - l.start
	return l.input[l.start:l.pos], true
}
func (l *lexer) backupTerm() {
	l.pos -= l.width
}
func (l *lexer) peekTerm() string {
	s := l.nextTerm()
	l.backupTerm()
	return s
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
	if l.nextTerm() == KeySelect {
		l.emit(itemSelect)
		return lexField
	}
	l.backupTerm()
	return l.errorf("syntax error: start with %q", l.input[l.pos])
}

func lexField(l *lexer) stateFunc {
	l.skipSpace()
	for {
		if s, ok := l.nextTermWithDot(); s != "" && ok {
			if agg, ok := Aggragation[s]; ok {
				l.emit(agg)
				if !l.accept(MarkLeftParen) {
					return l.errorf("syntax error: aggragation error, %q", l.input[l.pos:])
				}
				l.emit(itemLeftParen)
				if aggField, ok := l.nextTermWithDot(); aggField != "" && ok {
					l.emit(itemIdentifier)
					if !l.accept(MarkRightParen) {
						return l.errorf("syntax error: aggragation error, %q", l.input[l.pos:])
					}
					l.emit(itemRightParen)
				} else {
					return l.errorf("syntax error: aggragation error, %q", l.input[l.pos:])
				}
			} else {
				l.emit(itemIdentifier)
				l.skipSpace()
				if !l.accept(MakrComma) {
					break
				} else {
					l.emit(itemComma)
				}
			}
		} else {
			return l.errorf("syntax error: query field %q not valid", l.input[l.pos:])
		}
	}
	if l.peekTerm() == KeyFrom {
		return lexFrom
	}
	if l.peekTerm() == KeyWhere {
		return lexWhere
	}
	return nil
}

func lexFrom(l *lexer) stateFunc {
	l.nextTerm()
	l.emit(itemFrom)
	if table := l.nextTerm(); table != "" {
		l.emit(itemIdentifier)
		if l.peekTerm() == KeyWhere {
			return lexWhere
		} else if l.peek() == eof {
			return nil
		} else {
			return l.errorf("syntax error: %q", l.input[l.start:l.pos])
		}
	} else {
		return l.errorf("syntax error: table name %q not valid", l.input[l.start:l.pos])
	}
}

func lexWhere(l *lexer) stateFunc {
	l.nextTerm()
	l.emit(itemWhere)
	return lexLeftSide
}

func lexLeftSide(l *lexer) stateFunc {
	return nil
}
