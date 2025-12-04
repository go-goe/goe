package model

import (
	"context"
	"database/sql"

	"github.com/go-goe/goe/enum"
)

type Driver interface {
	MigrateContext(context.Context, *Migrator) error
	DropTable(schema, table string) error
	DropColumn(schema, table, column string) error
	RenameColumn(schema, table, oldColumn, newName string) error
	RenameTable(schema, table, newName string) error
	Init() error
	KeywordHandler(string) string
	NewConnection() Connection
	NewTransaction(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	Stats() sql.DBStats
	Close() error
	ErrorTranslator() func(err error) error
	Config
}

type Config interface {
	Name() string
	GetDatabaseConfig() *DatabaseConfig
}

type Logger interface {
	InfoContext(ctx context.Context, msg string, kv ...any)
	WarnContext(ctx context.Context, msg string, kv ...any)
	ErrorContext(ctx context.Context, msg string, kv ...any)
}

type Connection interface {
	ExecContext(ctx context.Context, query *Query) error
	QueryRowContext(ctx context.Context, query *Query) Row
	QueryContext(ctx context.Context, query *Query) (Rows, error)
}

type Transaction interface {
	Connection
	Commit() error
	Rollback() error
	SavePoint() (SavePoint, error)
}

type SavePoint interface {
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

type Attributer interface {
	Attribute(Body) Attribute
	GetField() any
}
