package query

import (
	"context"
	"log"

	"github.com/olauro/goe"
	"github.com/olauro/goe/wh"
)

type stateDelete struct {
	addrMap map[uintptr]goe.Field
	config  *goe.Config
	conn    goe.Connection
	builder *goe.Builder
	ctx     context.Context
	err     error
}

func createDeleteState(am map[uintptr]goe.Field, conn goe.Connection, c *goe.Config, ctx context.Context, d goe.Driver, e error) *stateDelete {
	return &stateDelete{addrMap: am, conn: conn, builder: goe.CreateBuilder(d), config: c, ctx: ctx, err: e}
}

func Remove[T any](db *goe.DB, table *T, value T) error {
	return RemoveContext(context.Background(), db, table, value)
}

func RemoveContext[T any](ctx context.Context, db *goe.DB, table *T, value T) error {
	pks, pksValue, err := getArgsPks(db.AddrMap, table, value)
	if err != nil {
		return err
	}

	s := DeleteContext(ctx, db, table)
	helperOperation(s.builder, s.addrMap, pks, pksValue)
	return s.Where()
}

// Delete uses [context.Background] internally;
// to specify the context, use [query.DeleteContext].
//
// # Example
func Delete[T any](db *goe.DB, table *T) *stateDelete {
	return DeleteContext(context.Background(), db, table)
}

// DeleteContext creates a delete state for table
func DeleteContext[T any](ctx context.Context, db *goe.DB, table *T) *stateDelete {
	ptrArgs, err := getArgsTable(db.AddrMap, table)

	var state *stateDelete
	if err != nil {
		state = createDeleteState(nil, nil, db.Config, ctx, nil, err)
		return state
	}
	state = createDeleteState(db.AddrMap, db.ConnPool, db.Config, ctx, db.Driver, err)
	//TODO: ptrArgs to goe.Fields
	state.builder.Args = ptrArgs
	return state
}

func (s *stateDelete) Where(Brs ...wh.Operator) error {
	if s.err != nil {
		return s.err
	}

	s.err = helperWhere(s.builder, s.addrMap, Brs...)
	if s.err != nil {
		return s.err
	}

	s.builder.BuildDelete(s.addrMap)
	s.err = s.builder.BuildSqlDelete()
	if s.err != nil {
		return s.err
	}

	Sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + Sql)
	}
	return handlerValues(s.conn, Sql, s.builder.ArgsAny, s.ctx)
}
