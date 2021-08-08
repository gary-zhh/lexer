package sql

import "errors"

type SqlType int

const (
	SqlSelect SqlType = iota
)

var (
	parseError = errors.New("syntax error")
	aggError   = errors.New("aggragation error")
)

type state int

const (
	stateStart state = iota
	stateSelect
	stateFromTable
	stateField
	stateCondition
	stateGroupBy
	stateOrderBy
	stateSort
	stateEnd
	stateError
)

type model struct {
	Type         SqlType // currently set to select
	TableName    string  // currently set to graph
	Fields       []string
	Aggragations Aggragation
	Conditions   Condition
}

type parse struct {
	*lexer
	model
	state
	error
}

func NewParse(text string) *parse {
	return &parse{
		lexer: lex("sql", text),
		model: model{
			Fields: make([]string, 0),
			Aggragations: Aggragation{
				Items: make([]aggItem, 0),
			},
		},
		state: stateStart,
	}
}
func (p *parse) switchState(i itemType) {
	switch i {
	case itemError:
		p.state = stateError
		p.error = parseError
	case itemSelect:
		p.state = stateField
	case itemWhere:
		p.state = stateCondition
	case itemGroupBy:
		p.state = stateGroupBy
	case itemOrderBy:
		p.state = stateOrderBy
	case itemEOF:
		p.state = stateEnd
	case itemFrom:
		p.state = stateFromTable
	}

}
func (p *parse) Generate() {
	for {
		switch p.state {
		case stateError:
			return
		case stateStart:
			i := p.lexer.nextItem()
			p.switchState(i.typ)
		case stateField:
			p.getFields()

		case stateFromTable:
			i := p.lexer.nextItem()
			if i.typ == itemError {
				p.switchState(i.typ)
				break
			}
			p.TableName = i.val
			next := p.lexer.nextItem()
			p.switchState(next.typ)
		case stateCondition:

		}
	}
}

func (p *parse) getFields() {
	var next item
	for {
		i := p.lexer.nextItem()
		if i.typ == itemError {
			p.switchState(i.typ)
			return
		}
		if i.typ == itemIdentifier {
			p.Fields = append(p.Fields, i.val)
		} else if i.typ > itemAggragation {
			if p.lexer.nextItem().typ == itemError { // eliminate the left paren
				p.switchState(i.typ)
				return
			}
			field := p.lexer.nextItem()
			if field.typ == itemError {
				p.switchState(i.typ)
				return
			}
			p.Aggragations.Items = append(p.Aggragations.Items, aggItem{Field: field.val, Agg: itemType2AggType[i.typ]})
			if p.lexer.nextItem().typ == itemError { // eliminate the right paren
				p.switchState(i.typ)
				return
			}
		}
		if next = p.lexer.nextItem(); next.typ != itemComma {
			break
		}
	}
	if err := p.checkAgg(); err != nil {
		p.state = stateError
		p.error = err
	}
	p.switchState(next.typ)
}

func (p *parse) checkAgg() error {
	return nil
}
