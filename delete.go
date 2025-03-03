package goe

import (
	"context"

	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/query"
)

type stateDelete struct {
	conn    Connection
	builder builder
	ctx     context.Context
	err     error
}

func Remove[T any](table *T, value T, tx ...Transaction) error {
	return RemoveContext(context.Background(), table, value, tx...)
}

func RemoveContext[T any](ctx context.Context, table *T, value T, tx ...Transaction) error {
	pks, valuesPks, err := getArgsPks(addrMap, table, value)
	if err != nil {
		return err
	}

	s := DeleteContext(ctx, table, tx...)

	brs := make([]query.Operation, 0, len(pks))
	brs = append(brs, query.Equals(&pks[0], valuesPks[0]))
	for i := 1; i < len(pks); i++ {
		brs = append(brs, query.And())
		brs = append(brs, query.Equals(&pks[i], valuesPks[i]))
	}

	return s.Where(brs...)
}

// Delete uses [context.Background] internally;
// to specify the context, use [query.DeleteContext].
//
// # Example
func Delete[T any](table *T, tx ...Transaction) *stateDelete {
	return DeleteContext(context.Background(), table, tx...)
}

// DeleteContext creates a delete state for table
func DeleteContext[T any](ctx context.Context, table *T, tx ...Transaction) *stateDelete {
	fields, err := getArgsTable(addrMap, table)

	var state *stateDelete
	if err != nil {
		state = new(stateDelete)
		state.err = ErrInvalidArg
		return state
	}

	db := fields[0].getDb()

	if tx != nil {
		state = createDeleteState(tx[0], ctx)
	} else {
		state = createDeleteState(db.Driver.NewConnection(), ctx)
	}

	state.builder.fields = fields
	return state
}

func (s *stateDelete) Where(Brs ...query.Operation) error {
	if s.err != nil {
		return s.err
	}

	s.err = helperWhere(&s.builder, addrMap, Brs...)
	if s.err != nil {
		return s.err
	}

	s.err = s.builder.buildSqlDelete()
	if s.err != nil {
		return s.err
	}

	return handlerValues(s.conn, s.builder.query, s.ctx)
}

func createDeleteState(conn Connection, ctx context.Context) *stateDelete {
	return &stateDelete{conn: conn, builder: createBuilder(enum.DeleteQuery), ctx: ctx}
}
