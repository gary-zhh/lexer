package sql

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	name       string
	input      string // the string being scanned
	start      int    // start position of this item
	pos        int    // current position in the input
	width      int    // width of last rune read
	parenDepth int
	items      chan item // channel of scanned items
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
	if l.pos >= len(l.input) {
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

// TODO: rewrite this function and add " as xxx "/ "as "xxx" "
// TODO: add "*" to indicating get all the fields
func lexField(l *lexer) stateFunc {
	l.skipSpace()
	for {
		if s, ok := l.nextTermWithDot(); s != "" && ok {
			if agg, ok := Aggragation[s]; ok {
				l.emit(agg)
				l.skipSpace()
				if !l.accept(MarkLeftParen) {
					return l.errorf("syntax error: aggragation error, %q", l.input[l.pos:])
				}
				l.emit(itemLeftParen)
				// TODO: take count(*) into consideration
				if aggField, ok := l.nextTermWithDot(); aggField != "" && ok {
					l.emit(itemIdentifier)
					l.skipSpace()
					if !l.accept(MarkRightParen) {
						return l.errorf("syntax error: aggragation error, %q", l.input[l.pos:])
					}
					l.emit(itemRightParen)
					l.skipSpace()
					if !l.accept(MakrComma) {
						break
					} else {
						l.emit(itemComma)
					}
				} else {
					return l.errorf("syntax error: aggragation error, %q", l.input[l.pos:])
				}
			} else {
				l.emit(itemIdentifier)
				l.skipSpace()
				// if n := l.nextTerm(); n == KeyAs {
				//
				// }
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
	return lexCondition
}

func lexCondition(l *lexer) stateFunc {
	l.skipSpace()
	switch r := l.next(); {
	case r == '(':
		l.emit(itemLeftParen)
		l.parenDepth++
	case r == ')':
		l.emit(itemRightParen)
		l.parenDepth--
		if l.parenDepth < 0 {
			return l.errorf("unexpected right paren")
		}
	default:
		l.backup()
		term := l.nextTerm()
		if term == KeyNot {
			l.emit(itemNot)
			return lexCondition
		} else {
			l.backupTerm()
		}
	}
	return lexLeftHandSide
}
func lexLeftHandSide(l *lexer) stateFunc {
	l.skipSpace()
	switch r := l.next(); {
	case r == '"':
		for n := l.next(); n != '"'; {
			if n == eof {
				l.errorf("upclosed string")
			}
		}
		l.emit(itemString)
	case r == '+' || r == '-' || '0' <= r && r <= '9':
		curDigits := digits
		l.backup()
		l.accept("+-")

		if l.accept("0") && l.accept("xX") {
			curDigits += "abcdefABCDEF"
		}

		l.acceptRun(curDigits)
		if l.accept(".") {
			l.acceptRun(curDigits)
		}

		if l.accept("eE") {
			l.accept("+-")
			l.acceptRun(digits)
		}
		l.emit(itemNumber)
	case unicode.IsLetter(r):
		l.backup()
		l.nextTermWithDot()
		l.emit(itemIdentifier)
	default:
		return l.errorf("syntax error: condition")
	}
	return lexCompare
}

func lexCompare(l *lexer) stateFunc {
	l.skipSpace()
	switch r := l.next(); {
	case r == '>':
		if l.accept("=") {
			l.emit(itemGreaterEqual)
		} else {
			l.emit(itemGreater)
		}
	case r == '<':
		if l.accept("=") {
			l.emit(itemLessEqual)
		} else {
			l.emit(itemLess)
		}
	case r == '!':
		if l.accept("=") {
			l.emit(itemNotEqual)
		} else {
			return l.errorf("syntax error: ")
		}
	case r == '=':
		l.emit(itemEqual)
	case r == 'l':
		l.backup()
		if n := l.nextTerm(); n == KeyLike {
			l.emit(itemLike)
		} else {
			return l.errorf("syntax error: ")
		}
	default:
		return l.errorf("syntax error: ")
	}
	return lexRightHandSide
}

func lexRightHandSide(l *lexer) stateFunc {
	l.skipSpace()
	switch r := l.next(); {
	case r == '"':
		for n := l.next(); n != '"'; {
			if n == eof {
				l.errorf("upclosed string")
			}
			n = l.next()
		}
		l.emit(itemString)
	case r == '+' || r == '-' || '0' <= r && r <= '9':
		curDigits := digits
		l.backup()
		l.accept("+-")

		if l.accept("0") && l.accept("xX") {
			curDigits += "abcdefABCDEF"
		}

		l.acceptRun(curDigits)
		if l.accept(".") {
			l.acceptRun(curDigits)
		}

		if l.accept("eE") {
			l.accept("+-")
			l.acceptRun(digits)
		}
		l.emit(itemNumber)
	case unicode.IsLetter(r):
		l.backup()
		l.nextTermWithDot()
		l.emit(itemIdentifier)
	default:
		return l.errorf("syntax error: condition")
	}
	return lexLogic
}

func lexLogic(l *lexer) stateFunc {
	l.skipSpace()
	for l.accept(")") {
		l.emit(itemRightParen)
		l.parenDepth--
		if l.parenDepth < 0 {
			return l.errorf("unexpected right paren")
		}
		l.skipSpace()
	}
	l.skipSpace()
	switch r := l.next(); {
	case r == 'a':
		l.backup()
		if n := l.nextTerm(); n != KeyAnd {
			return l.errorf("syntax error: ")
		}
		l.emit(itemAnd)
	case r == 'o':
		l.backup()
		n := l.nextTerm()
		if n == KeyOr {
			l.emit(itemOr)
		} else if n == "order" {
			if l.parenDepth != 0 {
				return l.errorf("syntax error: unclosed paren")
			}
			start := l.start
			if l.nextTerm() == "by" {
				l.start = start
				return lexOrderBy
			}
			return l.errorf("syntax error: ")
		} else {
			return l.errorf("syntax error: ")
		}

	case r == 'g':
		l.backup()
		// TODO : pos not correct
		if l.nextTerm() == "group" {
			if l.parenDepth != 0 {
				return l.errorf("syntax error: unclosed paren")
			}
			start := l.start
			if l.nextTerm() == "by" {
				l.start = start
				return lexGroupBy
			}
			return l.errorf("syntax error: ")
		} else {
			return l.errorf("syntax error: ")
		}
	case r == eof:
		return nil
	}
	return lexCondition
}

func lexGroupBy(l *lexer) stateFunc {
	l.emit(itemGroupBy)
	for {
		if s, ok := l.nextTermWithDot(); s != "" && ok {
			if agg, ok := Aggragation[s]; ok {
				l.emit(agg)
				l.skipSpace()
				if !l.accept(MarkLeftParen) {
					return l.errorf("syntax error: aggragation error, %q", l.input[l.pos:])
				}
				l.emit(itemLeftParen)
				if aggField, ok := l.nextTermWithDot(); aggField != "" && ok {
					l.emit(itemIdentifier)
					l.skipSpace()
					if !l.accept(MarkRightParen) {
						return l.errorf("syntax error: aggragation error, %q", l.input[l.pos:])
					}
					l.emit(itemRightParen)
					l.skipSpace()
					if !l.accept(MakrComma) {
						break
					} else {
						l.emit(itemComma)
					}
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
	n := l.nextTerm()
	if n == "order" {
		start := l.start
		if l.nextTerm() == "by" {
			l.start = start
			return lexOrderBy
		}
		return l.errorf("syntax error: ")
	}
	return lexCheckEnd
}

func lexOrderBy(l *lexer) stateFunc {
	l.emit(itemOrderBy)
	for {
		if s, ok := l.nextTermWithDot(); s != "" && ok {
			if agg, ok := Aggragation[s]; ok {
				l.emit(agg)
				l.skipSpace()
				if !l.accept(MarkLeftParen) {
					return l.errorf("syntax error: aggragation error, %q", l.input[l.pos:])
				}
				l.emit(itemLeftParen)
				if aggField, ok := l.nextTermWithDot(); aggField != "" && ok {
					l.emit(itemIdentifier)
					l.skipSpace()
					if !l.accept(MarkRightParen) {
						return l.errorf("syntax error: aggragation error, %q", l.input[l.pos:])
					}
					l.emit(itemRightParen)
					l.skipSpace()
					if !l.accept(MakrComma) {
						break
					} else {
						l.emit(itemComma)
					}
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
	n := l.nextTerm()
	if n == KeyDesc {
		return lexDesc
	} else if n == KeyAsc {
		return lexAsc
	}
	return lexCheckEnd
}

func lexDesc(l *lexer) stateFunc {
	l.emit(itemDesc)
	return lexCheckEnd
}

func lexAsc(l *lexer) stateFunc {
	l.emit(itemAsc)
	return lexCheckEnd
}

func lexCheckEnd(l *lexer) stateFunc {
	l.skipSpace()
	if l.pos >= len(l.input) {
		return nil
	}
	return l.errorf("syntax error: end with %q", l.input[l.pos:])
}
