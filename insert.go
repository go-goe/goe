package goe

import (
	"context"
	"errors"
	"reflect"

	"github.com/olauro/goe/enum"
)

type stateInsert[T any] struct {
	conn    Connection
	builder builder
	ctx     context.Context
	err     error
}

// Insert uses [context.Background] internally;
// to specify the context, use [query.InsertContext].
//
// # Example
func Insert[T any](table *T, tx ...Transaction) *stateInsert[T] {
	return InsertContext(context.Background(), table, tx...)
}

// InsertContext creates a insert state for table
func InsertContext[T any](ctx context.Context, table *T, tx ...Transaction) *stateInsert[T] {
	fields, err := getArgsTable(addrMap.mapField, table)

	var state *stateInsert[T]
	if err != nil {
		state = new(stateInsert[T])
		state.err = err
		return state
	}
	db := fields[0].getDb()

	if tx != nil {
		state = createInsertState[T](tx[0], ctx)
	} else {
		state = createInsertState[T](db.driver.NewConnection(), ctx)
	}
	state.builder.fields = fields
	return state
}

func (s *stateInsert[T]) One(value *T) error {
	if s.err != nil {
		return s.err
	}

	if value == nil {
		return errors.New("goe: invalid insert value. try sending a pointer to a struct as value")
	}

	v := reflect.ValueOf(value).Elem()

	pkFieldId := s.builder.buildSqlInsert(v)

	if s.builder.query.ReturningId != nil {
		return handlerValuesReturning(s.conn, s.builder.query, v, pkFieldId, s.ctx)
	}
	return handlerValues(s.conn, s.builder.query, s.ctx)
}

func (s *stateInsert[T]) All(value []T) error {
	if len(value) == 0 {
		return errors.New("goe: can't insert a empty batch value")
	}

	valueOf := reflect.ValueOf(value)

	pkFieldId := s.builder.buildSqlInsertBatch(valueOf)

	return handlerValuesReturningBatch(s.conn, s.builder.query, valueOf, pkFieldId, s.ctx)
}

func createInsertState[T any](conn Connection, ctx context.Context) *stateInsert[T] {
	return &stateInsert[T]{conn: conn, builder: createBuilder(enum.InsertQuery), ctx: ctx}
}

func getArgsTable[T any](AddrMap map[uintptr]field, table *T) ([]field, error) {
	if table == nil {
		return nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	fields := make([]field, 0)

	valueOf := reflect.ValueOf(table).Elem()
	if valueOf.Kind() != reflect.Struct {
		return nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	var fieldOf reflect.Value
	for i := 0; i < valueOf.NumField(); i++ {
		fieldOf = valueOf.Field(i)
		if fieldOf.Kind() == reflect.Slice && fieldOf.Type().Elem().Kind() == reflect.Struct {
			continue
		}
		addr := uintptr(fieldOf.Addr().UnsafePointer())
		if AddrMap[addr] != nil {
			fields = append(fields, AddrMap[addr])
		}
	}

	if len(fields) == 0 {
		return nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	return fields, nil
}
