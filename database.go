package goe

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sync"

	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/model"
)

var ErrInvalidArg = errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
var ErrTooManyTablesUpdate = errors.New("goe: invalid table. try sending arguments from the same table")

var ErrInvalidScan = errors.New("goe: invalid scan target. try sending an address to a struct, value or pointer for scan")
var ErrInvalidOrderBy = errors.New("goe: invalid order by target. try sending a pointer")

var ErrInvalidInsertValue = errors.New("goe: invalid insert value. try sending a pointer to a struct as value")
var ErrInvalidInsertBatchValue = errors.New("goe: invalid insert value. try sending a pointer to a slice of struct as value")
var ErrEmptyBatchValue = errors.New("goe: can't insert a empty batch value")
var ErrInvalidInsertPointer = errors.New("goe: invalid insert value. try sending a pointer as value")

var ErrInvalidInsertInValue = errors.New("goe: invalid insertIn value. try sending only two values or a size even slice")

var ErrInvalidUpdateValue = errors.New("goe: invalid update value. try sending a struct or a pointer to struct as value")

var ErrNotFound = errors.New("goe: not found any element on result set")

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
	Driver Driver
}

func (db *DB) Stats() sql.DBStats {
	return db.Driver.Stats()
}

func (db *DB) RawQueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	return db.Driver.NewConnection().QueryContext(ctx, model.Query{Type: enum.RawQuery, RawSql: query, Arguments: args})
}

func (db *DB) RawExecContext(ctx context.Context, query string, args ...any) error {
	return db.Driver.NewConnection().ExecContext(ctx, model.Query{Type: enum.RawQuery, RawSql: query, Arguments: args})
}

// NewTransaction creates a new Transaction using the specified database target.
// It sets the isolation level to sql.LevelSerializable by default.
// The dbTarget parameter should be a valid database connection or instance.
// If successful, it returns the created Transaction; otherwise, it returns an error.
//
// NewTransaction uses [context.Background] internally;
// to specify the context, use [goe.NewTransactionContext]
func NewTransaction(dbTarget any) (Transaction, error) {
	return NewTransactionContext(context.Background(), dbTarget, sql.LevelSerializable)
}

func NewTransactionContext(ctx context.Context, dbTarget any, isolation sql.IsolationLevel) (Transaction, error) {
	goeDb := reflect.ValueOf(dbTarget).Elem().FieldByName("DB").Interface().(*DB)
	return goeDb.Driver.NewTransaction(ctx, &sql.TxOptions{Isolation: isolation})
}

func Close(dbTarget any) error {
	goeDb := getDatabase(dbTarget)
	err := goeDb.Driver.Close()
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
