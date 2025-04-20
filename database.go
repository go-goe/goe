package goe

import (
	"context"
	"database/sql"
	"reflect"
	"sync"
	"time"

	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
)

type goeMap struct {
	mu       sync.Mutex
	mapField map[uintptr]field
}

func (am *goeMap) get(key uintptr) field {
	am.mu.Lock()
	defer am.mu.Unlock()
	return am.mapField[key]
}

func (am *goeMap) set(key uintptr, value field) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.mapField[key] = value
}

func (am *goeMap) delete(key uintptr) {
	am.mu.Lock()
	defer am.mu.Unlock()
	delete(am.mapField, key)
}

var addrMap *goeMap

type DB struct {
	driver Driver
}

// Return the database stats as [sql.DBStats].
func (db *DB) Stats() sql.DBStats {
	return db.driver.Stats()
}

// Get the database name; SQLite, PostgreSQL...
func (db *DB) Name() string {
	return db.driver.Name()
}

func (db *DB) RawQueryContext(ctx context.Context, rawSql string, args ...any) (Rows, error) {
	query := model.Query{Type: enum.RawQuery, RawSql: rawSql, Arguments: args}
	var rows Rows
	rows, query.Header.Err = wrapperQuery(ctx, db.driver.NewConnection(), &query)
	if query.Header.Err != nil {
		return nil, db.driver.GetDatabaseConfig().ErrorQueryHandler(ctx, query)
	}
	db.driver.GetDatabaseConfig().InfoHandler(ctx, query)
	return rows, nil
}

func (db *DB) RawExecContext(ctx context.Context, rawSql string, args ...any) error {
	query := model.Query{Type: enum.RawQuery, RawSql: rawSql, Arguments: args}
	query.Header.Err = wrapperExec(ctx, db.driver.NewConnection(), &query)
	if query.Header.Err != nil {
		return db.driver.GetDatabaseConfig().ErrorQueryHandler(ctx, query)
	}
	db.driver.GetDatabaseConfig().InfoHandler(ctx, query)
	return nil
}

// NewTransaction creates a new Transaction using the specified database target.
// It sets the isolation level to sql.LevelSerializable by default.
// The dbTarget parameter should be a valid database connection or instance.
// If successful, it returns the created Transaction; otherwise, it returns an error.
//
// NewTransaction uses [context.Background] internally;
// to specify the context, use [goe.NewTransactionContext]
func (db *DB) NewTransaction() (Transaction, error) {
	return db.NewTransactionContext(context.Background(), sql.LevelSerializable)
}

func (db *DB) NewTransactionContext(ctx context.Context, isolation sql.IsolationLevel) (Transaction, error) {
	t, err := db.driver.NewTransaction(ctx, &sql.TxOptions{Isolation: isolation})
	if err != nil {
		return nil, db.driver.GetDatabaseConfig().ErrorHandler(ctx, err)
	}
	return t, nil
}

// Closes the database connection.
func Close(dbTarget any) error {
	goeDb := getDatabase(dbTarget)
	err := goeDb.driver.Close()
	if err != nil {
		return goeDb.driver.GetDatabaseConfig().ErrorHandler(context.TODO(), err)
	}

	valueOf := reflect.ValueOf(dbTarget).Elem()

	for i := range valueOf.NumField() - 1 {
		fieldOf := valueOf.Field(i).Elem()
		for fieldId := range fieldOf.NumField() {
			addrMap.delete(uintptr(fieldOf.Field(fieldId).Addr().UnsafePointer()))
		}
	}

	return nil
}

// Database config used by all GOE drivers
type DatabaseConfig struct {
	Logger           Logger
	IncludeArguments bool          // include all arguments used on query
	QueryThreshold   time.Duration // query threshold to warning on slow queries
	databaseName     string
}

func (c DatabaseConfig) ErrorHandler(ctx context.Context, err error) error {
	if c.Logger != nil {
		c.Logger.ErrorContext(ctx, "error", "database", c.databaseName, "err", err)
	}
	return err
}

func (c DatabaseConfig) ErrorQueryHandler(ctx context.Context, query model.Query) error {
	if c.Logger == nil {
		return query.Header.Err
	}
	logs := make([]any, 0)
	logs = append(logs, "database", c.databaseName)
	if c.IncludeArguments {
		logs = append(logs, "arguments", query.Arguments)
	}
	logs = append(logs, "sql", query.RawSql)
	logs = append(logs, "err", query.Header.Err)

	c.Logger.ErrorContext(ctx, "error", logs...)
	return query.Header.Err
}

func (c DatabaseConfig) InfoHandler(ctx context.Context, query model.Query) {
	if c.Logger == nil {
		return
	}
	qr := query.Header.QueryDuration + query.Header.QueryBuild + query.Header.ModelBuild

	logs := make([]any, 0)
	logs = append(logs, "database", c.databaseName)
	logs = append(logs, "query_duration", qr)
	if c.IncludeArguments {
		logs = append(logs, "arguments", query.Arguments)
	}
	logs = append(logs, "sql", query.RawSql)

	if c.QueryThreshold != 0 && qr > c.QueryThreshold {
		c.Logger.WarnContext(ctx, "query_threshold", logs...)
		return
	}

	c.Logger.InfoContext(ctx, "query_runned", logs...)
}

func getDatabase(dbTarget any) *DB {
	valueOf := reflect.ValueOf(dbTarget).Elem()
	return valueOf.Field(valueOf.NumField() - 1).Interface().(*DB)
}
