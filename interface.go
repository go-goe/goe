package goe

import (
	"context"
	"database/sql"
	"time"
)

type Field interface {
	IsPrimaryKey() bool
	GetSelect() string
	Table() []byte
	BuildAttributeSelect(*Builder, int)
	BuildAttributeInsert(*Builder)
	WriteAttributeInsert(*Builder)
	BuildAttributeUpdate(*Builder)
}

type Driver interface {
	Name() string
	Migrate(*Migrator, Connection)
	Init(*DB)
	KeywordHandler(string) string
	Select() []byte
	From() []byte
	Returning([]byte) []byte
}

type Connection interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Close() error
}

type ConnectionPool interface {
	Connection
	Conn(ctx context.Context) (*sql.Conn, error)
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}
