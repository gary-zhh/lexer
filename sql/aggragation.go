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

var (
	itemType2AggType = map[itemType]AggType{
		itemCount:    AggCount,
		itemSum:      AggSum,
		itemAverage:  AggAverage,
		itemMin:      AggMin,
		itemMax:      AggMax,
		itemDistinct: AggDistinct,
	}
)

type aggItem struct {
	Agg   AggType
	Field string
}
