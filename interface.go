package goe

import (
	"context"
	"database/sql"
)

type field interface {
	fieldSelect
	fieldDb
	isPrimaryKey() bool
	getTableId() int
	getSelect() string
	getAttributeName() string
	table() string
	buildAttributeInsert(*builder)
	writeAttributeInsert(*builder)
	buildAttributeUpdate(*builder)
}

type fieldDb interface {
	getDb() *DB
}

type fieldSelect interface {
	fieldDb
	buildAttributeSelect(*builder)
}

type Driver interface {
	Name() string
	MigrateContext(context.Context, *Migrator) (string, error)
	DropTable(string) (string, error)
	DropColumn(table, column string) (string, error)
	RenameColumn(table, oldColumn, newColumn string) (string, error)
	Init()
	KeywordHandler(string) string
	NewConnection() Connection
	NewTransaction(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	Stats() sql.DBStats
	Sql
}

type Connection interface {
	ExecContext(ctx context.Context, query Query) error
	QueryRowContext(ctx context.Context, query Query) Row
	QueryContext(ctx context.Context, query Query) (Rows, error)
}

type Transaction interface {
	Connection
	Commit() error
	Rollback() error
}

type Rows interface {
	Close() error
	Next() bool
	Row
}

type Row interface {
	Scan(dest ...any) error
}

// TODO: Remove this
type Sql interface {
	Select() []byte
	From() []byte
	Where() []byte
	Insert() []byte
	Values() []byte
	Returning([]byte) []byte
	Update() []byte
	Set() []byte
	Delete() []byte
}

const (
	SelectQuery uint = iota
	InsertQuery
	UpdateQuery
	DeleteQuery
)

const (
	_                   = iota
	CountAggregate uint = 1
)

const (
	_                  = iota
	UpperFunction uint = 1
)

const (
	LogicalWhere uint = iota
	OperationWhere
	OperationArgumentWhere
	OperationIsWhere
)

type Attribute struct {
	Table         string
	Name          string
	AggregateType uint
	FunctionType  uint
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
	Type      uint
	Attribute Attribute
	Value     any
	Operator  string
	ValueFlag string
}

type OrderBy struct {
	Desc      bool
	Attribute Attribute
}

type Query struct {
	Type       uint
	Attributes []Attribute
	Tables     []string

	Joins   []Join   //Select
	Limit   uint     //Select
	Offset  uint     //Select
	OrderBy *OrderBy //Select

	WhereOperations []Where //Select, Update and Delete
	Arguments       []any   //Insert and Update

	ReturningId    *Attribute //Insert
	BatchSizeQuery int        //Insert
	SizeArguments  int        //Insert

	RawSql string
}
