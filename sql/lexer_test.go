package sql

import (
	"fmt"
	"testing"
)

type lexTest struct {
	input string
	items []item
}

var lexStartTest = []lexTest{
	{
		"select", []item{{itemSelect, "select"}},
	},
	{
		" select", []item{{itemSelect, "select"}},
	},
	{
		"  select", []item{{itemSelect, "select"}},
	},
	{
		"select  ", []item{{itemSelect, "select"}},
	},
	{
		"notselect", []item{{itemError, "syntax error: start with 'n'"}},
	},
}

func Test_LexStart(t *testing.T) {
	for _, i := range lexStartTest {
		l := &lexer{
			input: i.input,
			items: make(chan item),
		}
		go func() {
			lexStart(l)
		}()
		it := <-l.items
		if it.typ != i.items[0].typ || it.val != i.items[0].val {
			t.Error("not right:", it.typ, it.val)
		} else {
			fmt.Println(it.typ, it.val)
		}
	}
}
