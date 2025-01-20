package query

import (
	"context"
	"log"
	"reflect"

	"github.com/olauro/goe"
)

type stateInsert[T any] struct {
	config  *goe.Config
	conn    goe.Connection
	builder *goe.Builder
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
	fields, err := getArgsTable(goe.AddrMap, table)

	var state *stateInsert[T]
	if err != nil {
		state = createInsertState[T](nil, nil, ctx, nil, err)
		return state
	}
	db := fields[0].GetDb()
	state = createInsertState[T](db.ConnPool, db.Config, ctx, db.Driver, err)
	state.builder.Fields = fields
	return state
}

func (s *stateInsert[T]) One(value *T) error {
	if s.err != nil {
		return s.err
	}

	if value == nil {
		return goe.ErrInvalidInsertValue
	}

	v := reflect.ValueOf(value).Elem()

	s.builder.BuildInsert()
	idName := s.builder.BuildValues(v)

	sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + sql)
	}
	if s.builder.Returning != nil {
		return handlerValuesReturning(s.conn, sql, v, s.builder.ArgsAny, idName, s.ctx)
	}
	return handlerValues(s.conn, sql, s.builder.ArgsAny, s.ctx)
}

func (s *stateInsert[T]) All(value []T) error {
	if len(value) == 0 {
		return goe.ErrEmptyBatchValue
	}

	valueOf := reflect.ValueOf(value)

	s.builder.BuildInsert()
	idName := s.builder.BuildBatchValues(valueOf)

	Sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + Sql)
	}
	return handlerValuesReturningBatch(s.conn, Sql, valueOf, s.builder.ArgsAny, idName, s.ctx)
}

func createInsertState[T any](conn goe.Connection, c *goe.Config, ctx context.Context, d goe.Driver, e error) *stateInsert[T] {
	return &stateInsert[T]{conn: conn, builder: goe.CreateBuilder(d), config: c, ctx: ctx, err: e}
}

func getArgsTable[T any](AddrMap map[uintptr]goe.Field, table *T) ([]goe.Field, error) {
	if table == nil {
		return nil, goe.ErrInvalidArg
	}
	fields := make([]goe.Field, 0)

	valueOf := reflect.ValueOf(table).Elem()
	if valueOf.Kind() != reflect.Struct {
		return nil, goe.ErrInvalidArg
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
		return nil, goe.ErrInvalidArg
	}
	return fields, nil
}
