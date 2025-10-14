package enum

type WhereType uint

const (
	_ WhereType = iota
	LogicalWhere
	OperationWhere
	OperationAttributeWhere
	OperationIsWhere
	OperationInWhere
)

type QueryType uint

const (
	_ QueryType = iota
	SelectQuery
	InsertQuery
	UpdateQuery
	DeleteQuery
	RawQuery
)

type AggregateType uint

const (
	_ AggregateType = iota
	CountAggregate
	MaxAggregate
	MinAggregate
	SumAggregate
	AvgAggregate
)

type FunctionType uint

const (
	_ FunctionType = iota
	UpperFunction
	LowerFunction
)

type JoinType uint

const (
	_ JoinType = iota
	Join
	LeftJoin
	RightJoin
)

type OperatorType uint

const (
	_             OperatorType = iota
	Equals                     // =
	NotEquals                  // <>
	Is                         // IS
	IsNot                      // IS NOT
	Greater                    // >
	GreaterEquals              // >=
	Less                       // <
	LessEquals                 // <=
	In                         // IN
	NotIn                      // NOT IN
	Like                       // LIKE
	NotLike                    // NOT LIKE
	And                        // AND
	Or                         // OR
)
