package goe

import (
	"context"
	"database/sql"

	"github.com/go-goe/goe/model"
)

type field interface {
	fieldSelect
	fieldDb
	isPrimaryKey() bool
	getTableId() int
	getFieldId() int
	getAttributeName() string
	buildAttributeInsert(*builder)
}

type fieldDb interface {
	getDb() *DB
}

type fieldSelect interface {
	fieldDb
	buildAttributeSelect([]model.Attribute, int)
	table() string
	scheme() *string
	getTableId() int
}

type Driver interface {
	MigrateContext(context.Context, *Migrator) error
	DropTable(scheme, table string) error
	DropColumn(scheme, table, column string) error
	RenameColumn(scheme, table, oldColumn, newColumn string) error
	Init() error
	KeywordHandler(string) string
	NewConnection() Connection
	NewTransaction(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	Stats() sql.DBStats
	Close() error
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
	ExecContext(ctx context.Context, query *model.Query) error
	QueryRowContext(ctx context.Context, query *model.Query) Row
	QueryContext(ctx context.Context, query *model.Query) (Rows, error)
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
