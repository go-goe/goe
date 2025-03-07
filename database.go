package goe

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

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

var addrMap map[uintptr]field

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

func GetGoeDatabase(dbTarget any) (db *DB, err error) {
	dbValueOf := reflect.ValueOf(dbTarget).Elem()
	if dbValueOf.NumField() == 0 {
		return nil, fmt.Errorf("goe: Database %v with no structs", dbValueOf.Type().Name())
	}
	goeDb := addrMap[uintptr(dbValueOf.Field(0).UnsafePointer())]

	if goeDb == nil {
		return nil, fmt.Errorf("goe: Database %v with no structs", dbValueOf.Type().Name())
	}

	return goeDb.getDb(), nil
}

func NewTransaction(dbTarget any) (Transaction, error) {
	return NewTransactionContext(context.Background(), dbTarget, sql.LevelSerializable)
}

func NewTransactionContext(ctx context.Context, dbTarget any, isolation sql.IsolationLevel) (Transaction, error) {
	goeDb, err := GetGoeDatabase(dbTarget)
	if err != nil {
		return nil, err
	}

	return goeDb.Driver.NewTransaction(ctx, &sql.TxOptions{Isolation: isolation})
}
