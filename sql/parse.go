package sql

type SqlType int

const (
	SqlSelect SqlType = iota
)

type model struct {
	Type         SqlType // currently set to select
	TableName    string  // currently set to graph
	Fields       []string
	Aggragations []Aggragation
	Conditions   Condition
}

type parse struct {
	lexer *lexer
	model model
}

func NewParse(text string) *parse {
	return &parse{
		lexer: lex("sql", text),
		model: model{
			Fields:       make([]string, 0),
			Aggragations: make([]Aggragation, 0),
		},
	}
}

func (p *parse) Generate() error {

	return nil
}
