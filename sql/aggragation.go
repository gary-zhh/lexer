package sql

type Aggragation struct {
	Items []aggItem
}

type AggType int

const (
	AggCount AggType = iota
	AggSum
	AggAverage
	AggMin
	AggMax
	AggDistinct
)

type aggItem struct {
	Field string
	Agg   AggType
}
