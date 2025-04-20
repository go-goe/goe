package goe

import (
	"context"

	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
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

	return r.delete.Where(operations(pks, valuesPks))
}

// Delete remove records in the given table
//
// Delete uses [context.Background] internally;
// to specify the context, use [DeleteContext].
//
// # Examples
//
//	// delete all records
//	err = goe.Delete(db.UserRole).All()
//	// delete one record
//	err = goe.Delete(db.Animal).Where(where.Equals(&db.Animal.Id, 2))
func Delete[T any](table *T) *stateDelete {
	return DeleteContext(context.Background(), table)
}

// Delete remove records in the given table
//
// See [Delete] for examples
func DeleteContext[T any](ctx context.Context, table *T) *stateDelete {
	var state *stateDelete = createDeleteState(ctx)
	state.builder.fields = append(state.builder.fields, getArg(table, addrMap.mapField, nil))
	return state
}

func (s *stateDelete) OnTransaction(tx Transaction) *stateDelete {
	s.conn = tx
	return s
}

// Delete all records
func (s *stateDelete) All() error {
	return s.Where(model.Operation{})
}

// Where receives [model.Operation] as where operations from where sub package
func (s *stateDelete) Where(o model.Operation) error {
	if s.err != nil {
		return s.err
	}
	helperWhere(&s.builder, addrMap.mapField, o)

	s.builder.buildSqlDelete()

	driver := s.builder.fields[0].getDb().driver
	if s.conn == nil {
		s.conn = driver.NewConnection()
	}

	return handlerValues(s.ctx, s.conn, s.builder.query, driver.GetDatabaseConfig())
}

func createDeleteState(ctx context.Context) *stateDelete {
	return &stateDelete{builder: createBuilder(enum.DeleteQuery), ctx: ctx}
}
