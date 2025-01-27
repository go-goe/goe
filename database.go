package goe

import (
	"context"
	"database/sql"
	"errors"
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

type Config struct {
	LogQuery bool
}

var AddrMap map[uintptr]field

type DB struct {
	Config *Config
	SqlDB  *sql.DB
	Driver Driver
}

func BeginTx(dbTarget any) (*Tx, error) {
	return BeginTxContext(context.Background(), dbTarget, sql.LevelSerializable)
}

func BeginTxContext(ctx context.Context, dbTarget any, isolation sql.IsolationLevel) (*Tx, error) {
	goeDb, err := GetGoeDatabase(dbTarget)
	if err != nil {
		return nil, err
	}

	var sqlTx *sql.Tx
	sqlTx, err = goeDb.SqlDB.BeginTx(ctx, &sql.TxOptions{Isolation: isolation})
	if err != nil {
		return nil, err
	}

	return &Tx{SqlTx: sqlTx}, nil
}

type Tx struct {
	SqlTx *sql.Tx
}

func (tx *Tx) Commit() error {
	return tx.SqlTx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.SqlTx.Rollback()
}
