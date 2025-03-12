package model

import "github.com/olauro/goe/enum"

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
	Table          string
	FirstArgument  JoinArgument
	JoinOperation  string
	SecondArgument JoinArgument
}

type Where struct {
	Type           enum.WhereType
	Attribute      Attribute
	Operator       string
	AttributeValue Attribute
	SizeIn         uint
	QueryIn        *Query
}

type OrderBy struct {
	Desc      bool
	Attribute Attribute
}

type Query struct {
	Type       enum.QueryType
	Attributes []Attribute
	Tables     []string

	Joins   []Join   //Select
	Limit   uint     //Select
	Offset  uint     //Select
	OrderBy *OrderBy //Select

	WhereOperations []Where //Select, Update and Delete
	WhereIndex      int     //Start of where position arguments $1, $2...
	Arguments       []any

	ReturningId    *Attribute //Insert
	BatchSizeQuery int        //Insert
	SizeArguments  int        //Insert

	RawSql string
}
