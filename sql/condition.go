package sql

type ComparatorType int

const (
	ComparatorEQ ComparatorType = iota
	ComparatorNEQ
	ComparatorGT
	ComparatorGTE
	ComparatorLT
	ComparatorLTE
	ComparatorLIKE
)

type LogicType int

const (
	LogicAnd LogicType = iota
	LogicOr
	LogicNot
)

type Condition interface {
}

type SingleCondition struct {
	Field      string
	Comparator ComparatorType
	Value      interface{}
}
type MultiCondition struct {
	SubConditions []*Condition
	Logic         LogicType
}
