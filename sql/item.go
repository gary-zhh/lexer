package sql

import "fmt"

type item struct {
	typ itemType
	val string
}
type itemType int

const (
	itemError        itemType = iota // error occurred; value is text of error
	itemBool                         // boolean constant
	itemChar                         // printable ASCII character; grab bag for comma etc.
	itemCharConstant                 // character constant
	itemNumber                       // simple number, including imaginary
	itemIdentifier                   // alphanumeric identifier

	itemEqual        // "="
	itemGreater      // ">"
	itemGreaterEqual // ">="
	itemLess         // "<"
	itemLessEqual    // "<="
	itemNotEqual1    // "!="
	itemNotEqual2    // "<>"
	itemEOF

	itemComma      // ,
	itemLeftParen  // '(' inside action
	itemRawString  // raw quoted string (includes quotes) ``
	itemRightParen // ')' inside action
	itemSpace      // run of spaces separating arguments
	itemString     // quoted string (includes quotes)
	itemText       // plain text

	// Keywords appear after all the rest.
	itemKeyword // used only to delimit the keywords
	itemSelect
	itemAs
	itemFrom
	itemWhere
	itemGroupBy
	itemOrderBy
	itemAsc
	itemDesc
	itemLike
	itemAnd // and
	itemOr  // or
	itemNot // not
	itemCount
	itemAverage
	itemMax
	itemMin
	itemSum
	itemDistinct
)

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemKeyword:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

const (
	eof         = -1
	KeySelect   = "select"
	KeyFrom     = "from"
	KeyWhere    = "where"
	KeyNot      = "not"
	KeyAnd      = "and"
	KeyOr       = "or"
	KeyCount    = "count"
	KeyMax      = "max"
	KeyMin      = "min"
	KeySum      = "sum"
	KeyAverage  = "average"
	KeyDistinct = "distinct"
	Space       = " "

	MakrComma      = ","
	MarkDot        = "."
	MarkLeftParen  = "("
	MarkRightParen = ")"
)

var (
	Aggragation = map[string]itemType{
		KeyCount:    itemCount,
		KeyMax:      itemMax,
		KeyMin:      itemMin,
		KeySum:      itemSum,
		KeyAverage:  itemAverage,
		KeyDistinct: itemDistinct,
	}
	LogicOperator = map[string]itemType{
		KeyAnd: itemAnd,
		KeyOr:  itemOr,
		KeyNot: itemNot,
	}
	letter = "abcdefghijklmnopqrstuvwxyz"
	digits = "0123456789"
)
