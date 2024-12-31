package query

import (
	"context"
	"log"
	"reflect"

	"github.com/olauro/goe"
	"github.com/olauro/goe/wh"
)

type stateUpdate[T any] struct {
	config  *goe.Config
	conn    goe.Connection
	addrMap map[uintptr]goe.Field
	builder *goe.Builder
	ctx     context.Context
	err     error
}

func Save[T any, U any](db *goe.DB, table *T, pk *U, id U, value T) error {
	return SaveContext(context.Background(), db, table, pk, id, value)
}

func SaveContext[T any, U any](ctx context.Context, db *goe.DB, table *T, pk *U, id U, value T) error {
	if table == nil {
		return goe.ErrInvalidArg
	}
	includes, err := getArgsTable(db.AddrMap, table)

	if err != nil {
		return err
	}

	return UpdateContext(ctx, db, table).queryUpdate(includes, db.AddrMap).Where(wh.Equals(pk, id)).Value(value)
}

func Update[T any](db *goe.DB, table *T) *stateUpdate[T] {
	return UpdateContext[T](context.Background(), db, table)
}

// UpdateContext creates a update state for table
func UpdateContext[T any](ctx context.Context, db *goe.DB, table *T) *stateUpdate[T] {
	s := createUpdateState[T](db.AddrMap, db.ConnPool, db.Config, ctx, db.Driver, nil)
	if table == nil {
		s.err = goe.ErrInvalidArg
	}
	return s
}

func (s *stateUpdate[T]) Includes(args ...any) *stateUpdate[T] {
	if s.err != nil {
		return s
	}

	ptrArgs, err := getArgsUpdate(s.addrMap, args...)

	if err != nil {
		s.err = err
		return s
	}

	return s.queryUpdate(ptrArgs, s.addrMap)
}

func (s *stateUpdate[T]) Where(brs ...wh.Operator) *stateUpdate[T] {
	if s.err != nil {
		return s
	}
	s.err = helperWhere(s.builder, s.addrMap, brs...)
	return s
}

func (s *stateUpdate[T]) Value(value T) error {
	if s.err != nil {
		return s.err
	}

	if s.builder.Args == nil {
		//TODO: Includes error
		return goe.ErrInvalidArg
	}

	v := reflect.ValueOf(value)

	s.builder.BuildSet(v)

	//generate query
	s.err = s.builder.BuildSqlUpdate()
	if s.err != nil {
		return s.err
	}

	sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + sql)
	}
	return handlerValues(s.conn, sql, s.builder.ArgsAny, s.ctx)
}

func (s *stateUpdate[T]) queryUpdate(args []uintptr, addrMap map[uintptr]goe.Field) *stateUpdate[T] {
	if s.err == nil {
		s.builder.Args = append(s.builder.Args, args...)
		s.builder.BuildUpdate(addrMap)
	}
	return s
}

func getArgsUpdate(AddrMap map[uintptr]goe.Field, args ...any) ([]uintptr, error) {
	ptrArgs := make([]uintptr, 0)
	var valueOf reflect.Value
	for i := range args {
		valueOf = reflect.ValueOf(args[i])

		if valueOf.Kind() != reflect.Pointer {
			return nil, goe.ErrInvalidArg
		}

		valueOf = valueOf.Elem()
		addr := uintptr(valueOf.Addr().UnsafePointer())
		if AddrMap[addr] != nil {
			ptrArgs = append(ptrArgs, addr)
		}
	}
	if len(ptrArgs) == 0 {
		return nil, goe.ErrInvalidArg
	}
	return ptrArgs, nil
}

func createUpdateState[T any](
	am map[uintptr]goe.Field,
	conn goe.Connection, c *goe.Config,
	ctx context.Context,
	d goe.Driver,
	e error) *stateUpdate[T] {
	return &stateUpdate[T]{addrMap: am, conn: conn, builder: goe.CreateBuilder(d), config: c, ctx: ctx, err: e}
}
