package goe

import (
	"context"
	"database/sql"
	"reflect"
	"sync"

	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/model"
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

func (db *DB) Stats() sql.DBStats {
	return db.driver.Stats()
}

func (db *DB) RawQueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	return db.driver.NewConnection().QueryContext(ctx, model.Query{Type: enum.RawQuery, RawSql: query, Arguments: args})
}

func (db *DB) RawExecContext(ctx context.Context, query string, args ...any) error {
	return db.driver.NewConnection().ExecContext(ctx, model.Query{Type: enum.RawQuery, RawSql: query, Arguments: args})
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
	return db.driver.NewTransaction(ctx, &sql.TxOptions{Isolation: isolation})
}

func Close(dbTarget any) error {
	goeDb := getDatabase(dbTarget)
	err := goeDb.driver.Close()
	if err != nil {
		return err
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
