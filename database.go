package goe

import (
	"context"
	"database/sql"
	"reflect"
	"sync"

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
	driver model.Driver
}

// Return the database stats as [sql.DBStats].
func (db *DB) Stats() sql.DBStats {
	return db.driver.Stats()
}

// Get the database name; SQLite, PostgreSQL...
func (db *DB) Name() string {
	return db.driver.Name()
}

func (db *DB) RawQueryContext(ctx context.Context, rawSql string, args ...any) (model.Rows, error) {
	query := model.Query{Type: enum.RawQuery, RawSql: rawSql, Arguments: args}
	var rows model.Rows
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

// NewTransaction creates a new Transaction on the database.
// It sets the isolation level to sql.LevelSerializable by default.
//
// NewTransaction uses [context.Background] internally;
// to specify the context and the isolation level, use [NewTransactionContext]
func (db *DB) NewTransaction() (model.Transaction, error) {
	return db.NewTransactionContext(context.Background(), sql.LevelSerializable)
}

func (db *DB) NewTransactionContext(ctx context.Context, isolation sql.IsolationLevel) (model.Transaction, error) {
	t, err := db.driver.NewTransaction(ctx, &sql.TxOptions{Isolation: isolation})
	if err != nil {
		return nil, db.driver.GetDatabaseConfig().ErrorHandler(ctx, err)
	}
	return t, nil
}

func (db *DB) BeginTransaction(txFunc func(Transaction) error) error {
	return db.BeginTransactionContext(context.Background(), sql.LevelSerializable, txFunc)
}

func (db *DB) BeginTransactionContext(ctx context.Context, isolation sql.IsolationLevel, txFunc func(Transaction) error) (err error) {
	var t model.Transaction
	if t, err = db.NewTransactionContext(ctx, isolation); err != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			t.Rollback()
		}
	}()
	if err = txFunc(Transaction{t}); err != nil {
		t.Rollback()
		return
	}
	return t.Commit()
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

func getDatabase(dbTarget any) *DB {
	valueOf := reflect.ValueOf(dbTarget).Elem()
	return valueOf.Field(valueOf.NumField() - 1).Interface().(*DB)
}
