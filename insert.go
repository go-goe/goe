package goe

import (
	"context"
	"errors"
	"reflect"

	"github.com/go-goe/goe/enum"
)

type stateInsert[T any] struct {
	conn    Connection
	builder builder
	ctx     context.Context
}

// Insert inserts a new record into the given table.
//
// Insert uses [context.Background] internally;
// to specify the context, use [InsertContext].
//
// # Examples
//
//	// insert one record
//	err = goe.Insert(db.Person).One(&Person{Name: "Jhon"})
//	// insert a list of records
//	persons := []Person{{Name: "Jhon"}, {Name: "Mary"}}
//	err = goe.Insert(db.Person).All(persons)
func Insert[T any](table *T) stateInsert[T] {
	return InsertContext(context.Background(), table)
}

// InsertContext inserts a new record into the given table.
//
// See [Insert] for examples.
func InsertContext[T any](ctx context.Context, table *T) stateInsert[T] {
	var state stateInsert[T] = createInsertState[T](ctx)
	state.builder.fields = getArgsTable(addrMap.mapField, table)
	return state
}

func (s stateInsert[T]) OnTransaction(tx Transaction) stateInsert[T] {
	s.conn = tx
	return s
}

func (s stateInsert[T]) One(value *T) error {
	if value == nil {
		return errors.New("goe: invalid insert value. try sending a pointer to a struct as value")
	}

	v := reflect.ValueOf(value).Elem()

	pkFieldId := s.builder.buildSqlInsert(v)

	driver := s.builder.fields[0].getDb().driver
	if s.conn == nil {
		s.conn = driver.NewConnection()
	}

	if s.builder.query.ReturningId != nil {
		return handlerValuesReturning(s.ctx, s.conn, s.builder.query, v, pkFieldId, driver.GetDatabaseConfig())
	}
	return handlerValues(s.ctx, s.conn, s.builder.query, driver.GetDatabaseConfig())
}

func (s stateInsert[T]) All(value []T) error {
	if len(value) == 0 {
		return errors.New("goe: can't insert a empty batch value")
	}

	valueOf := reflect.ValueOf(value)

	pkFieldId := s.builder.buildSqlInsertBatch(valueOf)

	driver := s.builder.fields[0].getDb().driver
	if s.conn == nil {
		s.conn = driver.NewConnection()
	}

	return handlerValuesReturningBatch(s.ctx, s.conn, s.builder.query, valueOf, pkFieldId, driver.GetDatabaseConfig())
}

func createInsertState[T any](ctx context.Context) stateInsert[T] {
	return stateInsert[T]{builder: createBuilder(enum.InsertQuery), ctx: ctx}
}

func getArgsTable[T any](addrMap map[uintptr]field, table *T) []field {
	if table == nil {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	fields := make([]field, 0)

	valueOf := reflect.ValueOf(table).Elem()
	if valueOf.Kind() != reflect.Struct {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	var fieldOf reflect.Value
	for i := 0; i < valueOf.NumField(); i++ {
		fieldOf = valueOf.Field(i)
		if fieldOf.Kind() == reflect.Slice && fieldOf.Type().Elem().Kind() == reflect.Struct {
			continue
		}
		addr := uintptr(fieldOf.Addr().UnsafePointer())
		if addrMap[addr] != nil {
			fields = append(fields, addrMap[addr])
		}
	}

	if len(fields) == 0 {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	return fields
}
