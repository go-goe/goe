package goe

import (
	"context"
	"reflect"

	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
)

type stateDelete struct {
	conn    Connection
	builder builder
	ctx     context.Context
}

type remove[T any] struct {
	table  *T
	delete stateDelete
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
//	err = goe.Remove(db.Animal).ByID(Animal{Id: 2})
func Remove[T any](table *T) remove[T] {
	return RemoveContext(context.Background(), table)
}

func RemoveContext[T any](ctx context.Context, table *T) remove[T] {
	return remove[T]{table: table, delete: DeleteContext(ctx, table)}
}

func (r remove[T]) OnTransaction(tx Transaction) remove[T] {
	r.delete.conn = tx
	return r
}

func (r remove[T]) ByID(value T) error {
	pks, valuesPks, skip := getArgsRemove(getArgs{
		addrMap:   addrMap.mapField,
		tableArgs: getRemoveTableArgs(r.table),
		value:     value})

	// skip queries on empty models
	if skip {
		return nil
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
func Delete[T any](table *T) stateDelete {
	return DeleteContext(context.Background(), table)
}

// Delete remove records in the given table
//
// See [Delete] for examples
func DeleteContext[T any](ctx context.Context, table *T) stateDelete {
	var state stateDelete = createDeleteState(ctx)
	state.builder.fields = append(state.builder.fields, getArgDelete(table, addrMap.mapField))
	return state
}

func (s stateDelete) OnTransaction(tx Transaction) stateDelete {
	s.conn = tx
	return s
}

// Delete all records
func (s stateDelete) All() error {
	return s.Where(model.Operation{})
}

// Where receives [model.Operation] as where operations from where sub package
func (s stateDelete) Where(o model.Operation) error {
	helperWhere(&s.builder, addrMap.mapField, o)

	s.builder.buildSqlDelete()

	driver := s.builder.fields[0].getDb().driver
	if s.conn == nil {
		s.conn = driver.NewConnection()
	}

	return handlerValues(s.ctx, s.conn, s.builder.query, driver.GetDatabaseConfig())
}

func createDeleteState(ctx context.Context) stateDelete {
	return stateDelete{builder: createBuilder(enum.DeleteQuery), ctx: ctx}
}

func getArgsRemove(a getArgs) ([]any, []any, bool) {
	args, values := getPrimaryArgs(a)

	if len(args) == 0 {
		return nil, nil, true
	}
	return args, values, false
}

func getArgDelete(arg any, addrMap map[uintptr]field) field {
	v := reflect.ValueOf(arg)
	if v.Kind() != reflect.Pointer {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	addr := uintptr(v.UnsafePointer())
	if addrMap[addr] != nil {
		return addrMap[addr]
	}

	return nil
}

func getRemoveTableArgs(table any) []any {
	valueOf := reflect.ValueOf(table).Elem()

	if valueOf.Kind() != reflect.Struct {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	args := make([]any, 0, valueOf.NumField())
	var fieldOf reflect.Value
	for i := 0; i < valueOf.NumField(); i++ {
		fieldOf = valueOf.Field(i)
		if fieldOf.Kind() == reflect.Slice && fieldOf.Type().Elem().Kind() == reflect.Struct {
			continue
		}

		args = append(args, fieldOf.Addr().Interface())
	}

	return args
}
