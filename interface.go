package goe

import (
	"context"
	"database/sql"

	"github.com/olauro/goe/enum"
)

type field interface {
	fieldSelect
	fieldDb
	isPrimaryKey() bool
	getTableId() int
	getFieldId() int
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
	Init() error
	KeywordHandler(string) string
	NewConnection() Connection
	NewTransaction(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	Stats() sql.DBStats
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
	WhereIndex      int     //Start of where operations on slice arguments
	Arguments       []any

	ReturningId    *Attribute //Insert
	BatchSizeQuery int        //Insert
	SizeArguments  int        //Insert

	RawSql string
}
