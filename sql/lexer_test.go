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

var lexFieldTest = []lexTest{
	{
		"name, ", []item{{itemIdentifier, "name"}, {itemError, ""}},
	},
	{
		"name, age", []item{{itemIdentifier, "name"}, {itemError, ""}},
	},
	{
		"  name, count()", []item{{itemIdentifier, "name"}, {itemError, ""}},
	},
	{
		"  name  , sum(age)  ", []item{{itemIdentifier, "name"}, {itemError, ""}},
	},
}

func Test_LexField(t *testing.T) {
	for _, i := range lexFieldTest {
		l := &lexer{
			input: i.input,
			items: make(chan item),
		}
		go func() {
			for state := lexField; state != nil; {
				state = state(l)
			}
			close(l.items)
		}()
		for it := range l.items {
			fmt.Println(it)
		}
	}
}

func Test_LexCondition(t *testing.T) {
	l := &lexer{
		input: ``,
		items: make(chan item),
	}
	go func() {
		for state := lexCondition; state != nil; {
			state = state(l)
		}
		close(l.items)
	}()
	for it := range l.items {
		fmt.Println(it)
	}
}

func Test_All(t *testing.T) {
	l := lex("all", `select age, count(name) where (id= "1" and not(region = "cn-beijing")) group by region order by age`)
	for it := range l.items {
		fmt.Println(it)
	}
}
