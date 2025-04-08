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
	err     error
}

type create[T any] struct {
	table  *T
	tx     Transaction
	insert *stateInsert[T]
}

// Create is a wrapper over [Insert] for simple insert one record
// and return the inserted record with the new id.
//
// Create uses [context.Background] internally;
// to specify the context, use [CreateContext].
//
// # Examples
//
//	// create animal
//	insertedAnimal, err = goe.Create(db.Animal).ByValue(Animal{})
func Create[T any](table *T) *create[T] {
	return CreateContext(context.Background(), table)
}

func CreateContext[T any](ctx context.Context, table *T) *create[T] {
	return &create[T]{table: table, insert: InsertContext(ctx, table)}
}

func (c *create[T]) OnTransaction(tx Transaction) *create[T] {
	c.insert.OnTransaction(tx)
	c.tx = tx
	return c
}

func (c *create[T]) ByValue(value T) (*T, error) {
	err := c.insert.One(&value)
	if err != nil {
		return nil, err
	}

	return Find(c.table).OnTransaction(c.tx).ById(value)
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
func Insert[T any](table *T) *stateInsert[T] {
	return InsertContext(context.Background(), table)
}

// InsertContext inserts a new record into the given table.
//
// See [Insert] for examples.
func InsertContext[T any](ctx context.Context, table *T) *stateInsert[T] {
	fields, err := getArgsTable(addrMap.mapField, table)

	var state *stateInsert[T]
	if err != nil {
		state = new(stateInsert[T])
		state.err = err
		return state
	}
	state = createInsertState[T](ctx)

	state.builder.fields = fields
	return state
}

func (s *stateInsert[T]) OnTransaction(tx Transaction) *stateInsert[T] {
	s.conn = tx
	return s
}

func (s *stateInsert[T]) One(value *T) error {
	if s.err != nil {
		return s.err
	}

	if value == nil {
		return errors.New("goe: invalid insert value. try sending a pointer to a struct as value")
	}

	v := reflect.ValueOf(value).Elem()

	pkFieldId := s.builder.buildSqlInsert(v)

	if s.conn == nil {
		s.conn = s.builder.fields[0].getDb().driver.NewConnection()
	}

	if s.builder.query.ReturningId != nil {
		return handlerValuesReturning(s.conn, s.builder.query, v, pkFieldId, s.ctx)
	}
	return handlerValues(s.conn, s.builder.query, s.ctx)
}

func (s *stateInsert[T]) All(value []T) error {
	if len(value) == 0 {
		return errors.New("goe: can't insert a empty batch value")
	}

	valueOf := reflect.ValueOf(value)

	pkFieldId := s.builder.buildSqlInsertBatch(valueOf)

	if s.conn == nil {
		s.conn = s.builder.fields[0].getDb().driver.NewConnection()
	}

	return handlerValuesReturningBatch(s.conn, s.builder.query, valueOf, pkFieldId, s.ctx)
}

func createInsertState[T any](ctx context.Context) *stateInsert[T] {
	return &stateInsert[T]{builder: createBuilder(enum.InsertQuery), ctx: ctx}
}

func getArgsTable[T any](AddrMap map[uintptr]field, table *T) ([]field, error) {
	if table == nil {
		return nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	fields := make([]field, 0)

	valueOf := reflect.ValueOf(table).Elem()
	if valueOf.Kind() != reflect.Struct {
		return nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	var fieldOf reflect.Value
	for i := 0; i < valueOf.NumField(); i++ {
		fieldOf = valueOf.Field(i)
		if fieldOf.Kind() == reflect.Slice && fieldOf.Type().Elem().Kind() == reflect.Struct {
			continue
		}
		addr := uintptr(fieldOf.Addr().UnsafePointer())
		if AddrMap[addr] != nil {
			fields = append(fields, AddrMap[addr])
		}
	}

	if len(fields) == 0 {
		return nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	return fields, nil
}
