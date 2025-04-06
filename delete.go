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

type remove[T any] struct {
	table       *T
	tx          Transaction
	errNotFound error
	delete      *stateDelete
}

// Remove is a wrapper over [Delete] for more simple deletes,
// uses the value for create a where matching the primary keys.
// If the record don't exists returns a [ErrNotFound].
//
// Remove uses [context.Background] internally;
// to specify the context, use [RemoveContext].
//
// # Examples
//
//	// remove animal of id 2
//	err = goe.Remove(db.Animal).ById(Animal{Id: 2})
func Remove[T any](table *T) *remove[T] {
	return RemoveContext(context.Background(), table)
}

func RemoveContext[T any](ctx context.Context, table *T) *remove[T] {
	return &remove[T]{table: table, delete: DeleteContext(ctx, table), errNotFound: ErrNotFound}
}

func (r *remove[T]) OnTransaction(tx Transaction) *remove[T] {
	r.delete.OnTransaction(tx)
	r.tx = tx
	return r
}

// Replace the ErrNotFound with err
func (r *remove[T]) OnErrNotFound(err error) *remove[T] {
	r.errNotFound = err
	return r
}

func (r *remove[T]) ById(value T) error {
	pks, valuesPks, err := getArgsPks(getArgs{
		addrMap:     addrMap.mapField,
		table:       r.table,
		value:       value,
		errNotFound: r.errNotFound})

	if err != nil {
		return err
	}

	if _, err := Find(r.table).OnErrNotFound(r.errNotFound).OnTransaction(r.tx).ById(value); err != nil {
		return err
	}

	brs := make([]model.Operation, 0, len(pks))
	brs = append(brs, where.Equals(&pks[0], valuesPks[0]))
	for i := 1; i < len(pks); i++ {
		brs = append(brs, where.And())
		brs = append(brs, where.Equals(&pks[i], valuesPks[i]))
	}

	return r.delete.Wheres(brs...)
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
func Delete[T any](table *T) *stateDelete {
	return DeleteContext(context.Background(), table)
}

// Delete remove records in the given table
//
// See [Delete] for examples
func DeleteContext[T any](ctx context.Context, table *T) *stateDelete {
	field := getArg(table, addrMap.mapField, nil)

	var state *stateDelete
	if field == nil {
		state = new(stateDelete)
		state.err = errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
		return state
	}

	state = createDeleteState(ctx)

	state.builder.fields = append(state.builder.fields, field)
	return state
}

func (s *stateDelete) OnTransaction(tx Transaction) *stateDelete {
	s.conn = tx
	return s
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

	if s.conn == nil {
		s.conn = s.builder.fields[0].getDb().driver.NewConnection()
	}

	return handlerValues(s.conn, s.builder.query, s.ctx)
}

func createDeleteState(ctx context.Context) *stateDelete {
	return &stateDelete{builder: createBuilder(enum.DeleteQuery), ctx: ctx}
}
