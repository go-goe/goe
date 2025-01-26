package goe

import (
	"context"
	"log"
	"reflect"
)

type stateInsert[T any] struct {
	config  *Config
	conn    Connection
	builder *builder
	ctx     context.Context
	err     error
}

// Insert uses [context.Background] internally;
// to specify the context, use [query.InsertContext].
//
// # Example
func Insert[T any](table *T) *stateInsert[T] {
	return InsertContext[T](context.Background(), table)
}

// InsertContext creates a insert state for table
func InsertContext[T any](ctx context.Context, table *T) *stateInsert[T] {
	fields, err := getArgsTable(AddrMap, table)

	var state *stateInsert[T]
	if err != nil {
		state = createInsertState[T](nil, nil, ctx, nil, err)
		return state
	}
	db := fields[0].GetDb()
	state = createInsertState[T](db.ConnPool, db.Config, ctx, db.Driver, err)
	state.builder.fields = fields
	return state
}

func (s *stateInsert[T]) One(value *T) error {
	if s.err != nil {
		return s.err
	}

	if value == nil {
		return ErrInvalidInsertValue
	}

	v := reflect.ValueOf(value).Elem()

	s.builder.buildInsert()
	idName := s.builder.buildValues(v)

	sql := s.builder.sql.String()
	if s.config.LogQuery {
		log.Println("\n" + sql)
	}
	if s.builder.returning != nil {
		return handlerValuesReturning(s.conn, sql, v, s.builder.argsAny, idName, s.ctx)
	}
	return handlerValues(s.conn, sql, s.builder.argsAny, s.ctx)
}

func (s *stateInsert[T]) All(value []T) error {
	if len(value) == 0 {
		return ErrEmptyBatchValue
	}

	valueOf := reflect.ValueOf(value)

	s.builder.buildInsert()
	idName := s.builder.buildBatchValues(valueOf)

	Sql := s.builder.sql.String()
	if s.config.LogQuery {
		log.Println("\n" + Sql)
	}
	return handlerValuesReturningBatch(s.conn, Sql, valueOf, s.builder.argsAny, idName, s.ctx)
}

func createInsertState[T any](conn Connection, c *Config, ctx context.Context, d Driver, e error) *stateInsert[T] {
	return &stateInsert[T]{conn: conn, builder: createBuilder(d), config: c, ctx: ctx, err: e}
}

func getArgsTable[T any](AddrMap map[uintptr]Field, table *T) ([]Field, error) {
	if table == nil {
		return nil, ErrInvalidArg
	}
	fields := make([]Field, 0)

	valueOf := reflect.ValueOf(table).Elem()
	if valueOf.Kind() != reflect.Struct {
		return nil, ErrInvalidArg
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
		return nil, ErrInvalidArg
	}
	return fields, nil
}
