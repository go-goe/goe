package goe

import (
	"context"
	"log"

	"github.com/olauro/goe/query"
)

type stateDelete struct {
	config  *Config
	conn    Connection
	builder *builder
	ctx     context.Context
	err     error
}

func Remove[T any](table *T, value T, tx ...*Tx) error {
	return RemoveContext(context.Background(), table, value, tx...)
}

func RemoveContext[T any](ctx context.Context, table *T, value T, tx ...*Tx) error {
	pks, pksValue, err := getPksField(addrMap, table, value)
	if err != nil {
		return err
	}

	s := DeleteContext(ctx, table, tx...)
	helperOperation(s.builder, pks, pksValue)
	return s.Where()
}

// Delete uses [context.Background] internally;
// to specify the context, use [query.DeleteContext].
//
// # Example
func Delete[T any](table *T, tx ...*Tx) *stateDelete {
	return DeleteContext(context.Background(), table, tx...)
}

// DeleteContext creates a delete state for table
func DeleteContext[T any](ctx context.Context, table *T, tx ...*Tx) *stateDelete {
	fields, err := getArgsTable(addrMap, table)

	var state *stateDelete
	if err != nil {
		state = new(stateDelete)
		state.err = ErrInvalidArg
		return state
	}

	db := fields[0].getDb()

	if tx != nil {
		state = createDeleteState(tx[0].SqlTx, db.Config, ctx, db.Driver, err)
	} else {
		state = createDeleteState(db.SqlDB, db.Config, ctx, db.Driver, err)
	}

	state.builder.fields = fields
	return state
}

func (s *stateDelete) Where(Brs ...query.Operator) error {
	if s.err != nil {
		return s.err
	}

	s.err = helperWhere(s.builder, addrMap, Brs...)
	if s.err != nil {
		return s.err
	}

	s.err = s.builder.buildSqlDelete()
	if s.err != nil {
		return s.err
	}

	sql := s.builder.sql.String()
	if s.config.LogQuery {
		log.Println("\n" + sql)
	}
	return handlerValues(s.conn, sql, s.builder.argsAny, s.ctx)
}

func createDeleteState(conn Connection, c *Config, ctx context.Context, d Driver, e error) *stateDelete {
	return &stateDelete{conn: conn, builder: createBuilder(d), config: c, ctx: ctx, err: e}
}
