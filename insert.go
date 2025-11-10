package goe

import (
	"context"
	"errors"
	"reflect"

	"github.com/go-goe/goe/enum"
)

type stateInsert[T any] struct {
	conn    Connection
	table   *T
	builder builder
	ctx     context.Context
}

// Insert inserts a new record into the given table.
//
// Insert can return [ErrUniqueValue, ErrForeignKey and ErrBadRequest];
// use ErrBadRequest as a generic error for any user interaction.
//
// Insert uses [context.Background] internally;
// to specify the context, use [InsertContext].
//
// # Examples
//
//	// insert one record
//	err = goe.Insert(db.Person).One(&Person{Name: "John"})
//	// insert a list of records
//
//	persons := []Person{{Name: "John"}, {Name: "Mary"}}
//	err = goe.Insert(db.Person).All(persons)
func Insert[T any](table *T) stateInsert[T] {
	return InsertContext(context.Background(), table)
}

// InsertContext inserts a new record into the given table.
//
// See [Insert] for examples.
func InsertContext[T any](ctx context.Context, table *T) stateInsert[T] {
	var state stateInsert[T] = createInsertState(ctx, table)
	return state
}

// OnTransaction sets a transaction on the query.
//
// # Example
//
//	tx, err = db.NewTransaction()
//	if err != nil {
//		// handler error
//	}
//	defer tx.Rollback()
//
//	a := Animal{Name: "Cat"}
//	err = goe.Insert(db.Animal).OnTransaction(tx).One(&a)
//	if err != nil {
//		// handler error
//	}
//
//	err = tx.Commit()
//	if err != nil {
//		// handler error
//	}
func (s stateInsert[T]) OnTransaction(tx Transaction) stateInsert[T] {
	s.conn = tx
	return s
}

func (s stateInsert[T]) One(value *T) error {
	if value == nil {
		return errors.New("goe: invalid insert value. try sending a pointer to a struct as value")
	}
	valueOf := reflect.ValueOf(value).Elem()

	s.builder.fields = getArgsTable(addrMap.mapField, s.table, valueOf)

	pkFieldId := s.builder.buildSqlInsert(valueOf)

	driver := s.builder.fields[0].getDb().driver
	if s.conn == nil {
		s.conn = driver.NewConnection()
	}

	if s.builder.query.ReturningId != nil {
		return handlerValuesReturning(s.ctx, s.conn, s.builder.query, valueOf, pkFieldId, driver.GetDatabaseConfig())
	}
	return handlerValues(s.ctx, s.conn, s.builder.query, driver.GetDatabaseConfig())
}

func (s stateInsert[T]) All(value []T) error {
	if len(value) == 0 {
		return errors.New("goe: can't insert a empty batch value")
	}
	valueOf := reflect.ValueOf(value)

	s.builder.fields = getArgsTable(addrMap.mapField, s.table, valueOf)

	pkFieldId := s.builder.buildSqlInsertBatch(valueOf)

	driver := s.builder.fields[0].getDb().driver
	if s.conn == nil {
		s.conn = driver.NewConnection()
	}

	return handlerValuesReturningBatch(s.ctx, s.conn, s.builder.query, valueOf, pkFieldId, driver.GetDatabaseConfig())
}

func createInsertState[T any](ctx context.Context, t *T) stateInsert[T] {
	return stateInsert[T]{builder: createBuilder(enum.InsertQuery), ctx: ctx, table: t}
}

func getArgsTable(addrMap map[uintptr]field, table any, valueOf reflect.Value) []field {
	if table == nil {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	fields := make([]field, 0)

	tableValueOf := reflect.ValueOf(table).Elem()
	if tableValueOf.Kind() != reflect.Struct {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	var fieldOf reflect.Value
	for i := 0; i < tableValueOf.NumField(); i++ {
		fieldOf = tableValueOf.Field(i)
		if fieldOf.Kind() == reflect.Slice && fieldOf.Type().Elem().Kind() == reflect.Struct {
			continue
		}
		field := addrMap[uintptr(fieldOf.Addr().UnsafePointer())]
		if field != nil {
			if field.getDefault() && valueOf.Field(field.getFieldId()).IsZero() {
				continue
			}
			fields = append(fields, field)
		}
	}

	if len(fields) == 0 {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	return fields
}
