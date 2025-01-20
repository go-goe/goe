package query

import (
	"context"
	"log"

	"github.com/olauro/goe"
	"github.com/olauro/goe/wh"
)

type stateDelete struct {
	config  *goe.Config
	conn    goe.Connection
	builder *goe.Builder
	ctx     context.Context
	err     error
}

func createDeleteState(conn goe.Connection, c *goe.Config, ctx context.Context, d goe.Driver, e error) *stateDelete {
	return &stateDelete{conn: conn, builder: goe.CreateBuilder(d), config: c, ctx: ctx, err: e}
}

func Remove[T any](table *T, value T) error {
	return RemoveContext(context.Background(), table, value)
}

func RemoveContext[T any](ctx context.Context, table *T, value T) error {
	pks, pksValue, err := getArgsPks(goe.AddrMap, table, value)
	if err != nil {
		return err
	}

	s := DeleteContext(ctx, table)
	helperOperation(s.builder, pks, pksValue)
	return s.Where()
}

// Delete uses [context.Background] internally;
// to specify the context, use [query.DeleteContext].
//
// # Example
func Delete[T any](table *T) *stateDelete {
	return DeleteContext(context.Background(), table)
}

// DeleteContext creates a delete state for table
func DeleteContext[T any](ctx context.Context, table *T) *stateDelete {
	fields, err := getArgsTable(goe.AddrMap, table)

	var state *stateDelete
	if err != nil {
		state = createDeleteState(nil, nil, ctx, nil, err)
		return state
	}

	db := fields[0].GetDb()
	state = createDeleteState(db.ConnPool, db.Config, ctx, db.Driver, err)
	state.builder.Fields = fields
	return state
}

func (s *stateDelete) Where(Brs ...wh.Operator) error {
	if s.err != nil {
		return s.err
	}

	s.err = helperWhere(s.builder, goe.AddrMap, Brs...)
	if s.err != nil {
		return s.err
	}

	s.builder.BuildDelete()
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
