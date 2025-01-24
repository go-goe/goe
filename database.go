package goe

import (
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

var AddrMap map[uintptr]Field

type DB struct {
	Config   *Config
	ConnPool *sql.DB
	Driver   Driver
}
