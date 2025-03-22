package goe

import (
	"context"
	"errors"

	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/model"
	"github.com/olauro/goe/query/where"
)

type stateDelete struct {
	conn    Connection
	builder builder
	ctx     context.Context
	err     error
}

// Remove is a wrapper over [Delete] for more simple deletes,
// uses the value for create a where matching the primary keys.
//
// Remove uses [context.Background] internally;
// to specify the context, use [RemoveContext].
//
// # Examples
//
//	// remove animal of id 2
//	err = goe.Remove(db.Animal, Animal{Id: 2})
func Remove[T any](table *T, value T, tx ...Transaction) error {
	return RemoveContext(context.Background(), table, value, tx...)
}

func RemoveContext[T any](ctx context.Context, table *T, value T, tx ...Transaction) error {
	pks, valuesPks, err := getArgsPks(addrMap.mapField, table, value)
	if err != nil {
		return err
	}

	s := DeleteContext(ctx, table, tx...)

	brs := make([]model.Operation, 0, len(pks))
	brs = append(brs, where.Equals(&pks[0], valuesPks[0]))
	for i := 1; i < len(pks); i++ {
		brs = append(brs, where.And())
		brs = append(brs, where.Equals(&pks[i], valuesPks[i]))
	}

	return s.Wheres(brs...)
}

// Delete remove records in the given table
//
// Delete uses [context.Background] internally;
// to specify the context, use [DeleteContext].
//
// # Examples
//
//	// delete all records
//	err = goe.Delete(db.UserRole).Wheres()
//	// delete one record
//	err = goe.Delete(db.Animal).Wheres(where.Equals(&db.Animal.Id, 2))
func Delete[T any](table *T, tx ...Transaction) *stateDelete {
	return DeleteContext(context.Background(), table, tx...)
}

// Delete remove records in the given table
//
// See [Delete] for examples
func DeleteContext[T any](ctx context.Context, table *T, tx ...Transaction) *stateDelete {
	fields, err := getArgsTable(addrMap.mapField, table)

	var state *stateDelete
	if err != nil {
		state = new(stateDelete)
		state.err = errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
		return state
	}

	db := fields[0].getDb()

	if tx != nil {
		state = createDeleteState(tx[0], ctx)
	} else {
		state = createDeleteState(db.driver.NewConnection(), ctx)
	}

	state.builder.fields = fields
	return state
}

// Wheres receives [model.Operation] as where operations from where sub package
func (s *stateDelete) Wheres(brs ...model.Operation) error {
	if s.err != nil {
		return s.err
	}

	s.err = helperWhere(&s.builder, addrMap.mapField, brs...)
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
