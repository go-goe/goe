package query

import (
	"context"
	"log"
	"reflect"

	"github.com/olauro/goe"
)

type stateInsert[T any] struct {
	addrMap map[uintptr]goe.Field
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
func Insert[T any](db *goe.DB, table *T) *stateInsert[T] {
	return InsertContext[T](context.Background(), db, table)
}

// InsertContext creates a insert state for table
func InsertContext[T any](ctx context.Context, db *goe.DB, table *T) *stateInsert[T] {
	ptrArgs, err := getArgsTable(db.AddrMap, table)

	var state *stateInsert[T]
	if err != nil {
		state = createInsertState[T](nil, nil, db.Config, ctx, nil, err)
		return state
	}

	state = createInsertState[T](db.AddrMap, db.ConnPool, db.Config, ctx, db.Driver, err)
	state.builder.Args = ptrArgs
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

	s.builder.BuildInsert(s.addrMap)
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

	s.builder.BuildInsert(s.addrMap)
	idName := s.builder.BuildBatchValues(valueOf)

	Sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + Sql)
	}
	return handlerValuesReturningBatch(s.conn, Sql, valueOf, s.builder.ArgsAny, idName, s.ctx)
}

func createInsertState[T any](am map[uintptr]goe.Field, conn goe.Connection, c *goe.Config, ctx context.Context, d goe.Driver, e error) *stateInsert[T] {
	return &stateInsert[T]{addrMap: am, conn: conn, builder: goe.CreateBuilder(d), config: c, ctx: ctx, err: e}
}

func getArgsTable[T any](AddrMap map[uintptr]goe.Field, table *T) ([]uintptr, error) {
	if table == nil {
		return nil, goe.ErrInvalidArg
	}
	args := make([]uintptr, 0)

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
			args = append(args, addr)
		}
	}

	if len(args) == 0 {
		return nil, goe.ErrInvalidArg
	}
	return args, nil
}
