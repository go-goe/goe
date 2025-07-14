package model

import (
	"time"

	"github.com/go-goe/goe/enum"
)

type Attribute struct {
	Table         string
	Name          string
	AggregateType enum.AggregateType
	FunctionType  enum.FunctionType
}

type JoinArgument struct {
	Table string
	Name  string
}

type Join struct {
	Table          Table
	FirstArgument  JoinArgument
	JoinOperation  enum.JoinType
	SecondArgument JoinArgument
}

type Where struct {
	Type           enum.WhereType
	Attribute      Attribute
	Operator       enum.OperatorType
	AttributeValue Attribute
	SizeIn         uint
	QueryIn        *Query
}

type OrderBy struct {
	Desc      bool
	Attribute Attribute
}

type Table struct {
	Schema *string
	Name   string
}

func (t Table) String() string {
	if t.Schema != nil {
		return *t.Schema + "." + t.Name
	}
	return t.Name
}

type Query struct {
	Type       enum.QueryType
	Attributes []Attribute
	Tables     []Table

	Joins   []Join   //Select
	Limit   int      //Select
	Offset  int      //Select
	OrderBy *OrderBy //Select

	WhereOperations []Where //Select, Update and Delete
	WhereIndex      int     //Start of where position arguments $1, $2...
	Arguments       []any

	ReturningId    *Attribute //Insert
	BatchSizeQuery int        //Insert
	SizeArguments  int        //Insert

	RawSql string
	Header QueryHeader
}

type QueryHeader struct {
	Err           error
	ModelBuild    time.Duration
	QueryDuration time.Duration
}

type Operation struct {
	Type                enum.WhereType
	Arg                 any
	Value               ValueOperation
	Operator            enum.OperatorType
	Attribute           string
	Table               Table
	TableId             int
	Function            enum.FunctionType
	AttributeValue      string
	AttributeValueTable Table
	AttributeTableId    int
	FirstOperation      *Operation
	SecondOperation     *Operation
}

type Set struct {
	Attribute any
	Value     any
}

type Joins interface {
	FirstArg() any
	Join() enum.JoinType
	SecondArg() any
}

type Aggregate interface {
	Aggregate() enum.AggregateType
}

type FunctionType interface {
	GetType() enum.FunctionType
}

type ValueOperation interface {
	GetValue() any
}
