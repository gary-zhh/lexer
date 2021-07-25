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
	itemAndChar      // &&
	itemOrChar       // or
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
	itemAnd
	itemOr
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
	eof       = -1
	KeySelect = "select"
	KeyFrom   = "from"
	KeyWhere  = "where"
	Space     = " "

	MakrComma = ","
)

var (
	Aggragation = map[string]itemType{
		"count":    itemCount,
		"max":      itemMax,
		"min":      itemMin,
		"sum":      itemSum,
		"average":  itemAverage,
		"distinct": itemDistinct,
	}

	letter = "abcdefghijklmnopqrstuvwxyz"
)
