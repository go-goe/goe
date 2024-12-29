package goe

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/olauro/goe/jn"
	"github.com/olauro/goe/wh"
)

var ErrInvalidScan = errors.New("goe: invalid scan target. try sending an address to a struct, value or pointer for scan")
var ErrInvalidOrderBy = errors.New("goe: invalid order by target. try sending a pointer")

var ErrInvalidInsertValue = errors.New("goe: invalid insert value. try sending a pointer to a struct as value")
var ErrInvalidInsertBatchValue = errors.New("goe: invalid insert value. try sending a pointer to a slice of struct as value")
var ErrEmptyBatchValue = errors.New("goe: can't insert a empty batch value")
var ErrInvalidInsertPointer = errors.New("goe: invalid insert value. try sending a pointer as value")

var ErrInvalidInsertInValue = errors.New("goe: invalid insertIn value. try sending only two values or a size even slice")

var ErrInvalidUpdateValue = errors.New("goe: invalid update value. try sending a struct or a pointer to struct as value")

type stateSelect struct {
	config  *Config
	conn    Connection
	addrMap map[uintptr]Field
	builder *Builder
	ctx     context.Context
	err     error
}

func createSelectState(conn Connection, c *Config, ctx context.Context, d Driver, e error) *stateSelect {
	return &stateSelect{conn: conn, builder: CreateBuilder(d), config: c, ctx: ctx, err: e}
}

// Where creates a where SQL using the operations
func (s *stateSelect) Where(Brs ...wh.Operator) *stateSelect {
	if s.err != nil {
		return s
	}
	s.err = helperWhere(s.builder, s.addrMap, Brs...)
	return s
}

// Take takes i elements
//
// # Example
//
//	// takes frist 20 elements
//	db.Select(db.Habitat).Take(20)
//
//	// skips 20 and takes next 20 elements
//	db.Select(db.Habitat).Skip(20).Take(20).Scan(&h)
func (s *stateSelect) Take(i uint) *stateSelect {
	s.builder.Limit = i
	return s
}

// Skip skips i elements
//
// # Example
//
//	// skips frist 20 elements
//	db.Select(db.Habitat).Skip(20)
//
//	// skips 20 and takes next 20 elements
//	db.Select(db.Habitat).Skip(20).Take(20).Scan(&h)
func (s *stateSelect) Skip(i uint) *stateSelect {
	s.builder.Offset = i
	return s
}

// Page returns page p with i elements
//
// # Example
//
//	// returns first 20 elements
//	db.Select(db.Habitat).Page(1, 20).Scan(&h)
func (s *stateSelect) Page(p uint, i uint) *stateSelect {
	s.builder.Offset = i * (p - 1)
	s.builder.Limit = i
	return s
}

// OrderByAsc makes a ordained by arg ascending query
//
// # Example
//
//	// select first page of habitats orderning by name
//	db.Select(db.Habitat).Page(1, 20).OrderByAsc(&db.Habitat.Name).Scan(&h)
//
//	// same query
//	db.Select(db.Habitat).OrderByAsc(&db.Habitat.Name).Page(1, 20).Scan(&h)
func (s *stateSelect) OrderByAsc(arg any) *stateSelect {
	Field := getArg(arg, s.addrMap)
	if Field == nil {
		s.err = ErrInvalidOrderBy
		return s
	}
	s.builder.OrderBy = fmt.Sprintf("\nORDER BY %v ASC", Field.GetSelect())
	return s
}

// OrderByDesc makes a ordained by arg descending query
//
// # Example
//
//	// select last inserted habitat
//	db.Select(db.Habitat).Take(1).OrderByDesc(&db.Habitat.Id).Scan(&h)
//
//	// same query
//	db.Select(db.Habitat).OrderByDesc(&db.Habitat.Id).Take(1).Scan(&h)
func (s *stateSelect) OrderByDesc(arg any) *stateSelect {
	Field := getArg(arg, s.addrMap)
	if Field == nil {
		s.err = ErrInvalidOrderBy
		return s
	}
	s.builder.OrderBy = fmt.Sprintf("\nORDER BY %v DESC", Field.GetSelect())
	return s
}

func (s *stateSelect) querySelect(Args []uintptr, aggregates []aggregate) *stateSelect {
	if s.err == nil {
		s.builder.Args = Args
		s.builder.Aggregates = aggregates
		s.builder.BuildSelect(s.addrMap)
	}
	return s
}

// TODO: Add Doc
func (s *stateSelect) From(Tables ...any) *stateSelect {
	s.builder.Tables = make([]string, len(Tables))
	Args, err := getArgsTables(s.addrMap, s.builder.Tables, Tables...)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.Froms = Args
	return s
}

// TODO: Add Doc
func (s *stateSelect) Joins(joins ...jn.Joins) *stateSelect {
	if s.err != nil {
		return s
	}

	for _, j := range joins {
		Args, err := getArgsIn(s.addrMap, j.FirstArg(), j.SecondArg())
		if err != nil {
			s.err = err
			return s
		}
		s.builder.BuildSelectJoins(s.addrMap, j.Join(), Args)
	}
	return s
}

// Scan fills the target with the returned Sql data,
// target can be a pointer or a pointer to [Slice].
//
// In case of passing a pointer of struct or a pointer to slice of
// struct, goe package will match the Fields by name
//
// Scan uses [Sql.Row] if a not slice pointer is the target, in
// this case can return [Sql.ErrNoRows]
//
// Scan returns the SQL generated and a nil error if succeed.
//
// # Example:
//
//	// using struct
//	var a Animal
//	db.Select(db.Animal).Scan(&a)
//
//	// using slice
//	var a []Animal
//	db.Select(db.Animal).Scan(&a)
func (s *stateSelect) Scan(target any) error {
	if s.err != nil {
		return s.err
	}

	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		return ErrInvalidScan
	}
	value = value.Elem()
	if !value.CanSet() {
		return ErrInvalidScan
	}
	if value.Kind() == reflect.Ptr {
		value.Set(reflect.New(value.Type().Elem()))
		value = value.Elem()
	}

	//generate query
	s.err = s.builder.BuildSqlSelect()
	if s.err != nil {
		return s.err
	}

	Sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + Sql)
	}
	return handlerResult(s.conn, Sql, value, s.builder.ArgsAny, s.builder.StructColumns, s.builder.Limit, s.ctx)
}

/*
State Insert
*/
type stateInsert struct {
	config  *Config
	conn    Connection
	builder *Builder
	ctx     context.Context
	err     error
}

func createInsertState(conn Connection, c *Config, ctx context.Context, d Driver, e error) *stateInsert {
	return &stateInsert{conn: conn, builder: CreateBuilder(d), config: c, ctx: ctx, err: e}
}

func (s *stateInsert) queryInsert(Args []uintptr, addrMap map[uintptr]Field) *stateInsert {
	if s.err == nil {
		s.builder.Args = Args
		s.builder.buildInsert(addrMap)
	}
	return s
}

// Value inserts the value inside the database, and updates the Id Field if
// is a auto increment.
//
// The value needs to be a pointer to a struct of database types
// or a pointer to slice of database types (in case of batch insert).
//
// Value returns the SQL generated and error as nil if insert with success.
//
// # Example
//
//	// insert one value
//	food := Food{Id: "fc1865b4-6f2d-4cc6-b766-49c2634bf5c4", Name: "Cookie", Emoji: "🍪"}
//	db.Insert(db.Food).Value(&food)
//
//	// insert batch values
//	foods := []Food{
//		{Id: "401b5e23-5aa7-435e-ba4d-5c1b2f123596", Name: "Meat", Emoji: "🥩"},
//		{Id: "f023a4e7-34e9-4db2-85e0-efe8d67eea1b", Name: "Hotdog", Emoji: "🌭"},
//		{Id: "fc1865b4-6f2d-4cc6-b766-49c2634bf5c4", Name: "Cookie", Emoji: "🍪"},
//	}
//	db.Insert(db.Food).Value(&foods)
func (s *stateInsert) Value(value any) error {
	if s.err != nil {
		return s.err
	}

	v := reflect.ValueOf(value)

	if v.Kind() != reflect.Ptr {
		return ErrInvalidInsertPointer
	}

	v = v.Elem()

	if v.Kind() == reflect.Slice {
		return s.batchValue(v)
	}

	if v.Kind() != reflect.Struct {
		return ErrInvalidInsertValue
	}

	idName := s.builder.buildValues(v)

	Sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + Sql)
	}
	if s.builder.Returning != nil {
		return handlerValuesReturning(s.conn, Sql, v, s.builder.ArgsAny, idName, s.ctx)
	}
	return handlerValues(s.conn, Sql, s.builder.ArgsAny, s.ctx)
}

func (s *stateInsert) batchValue(value reflect.Value) error {
	if value.Len() == 0 {
		return ErrEmptyBatchValue
	}

	if value.Index(0).Kind() != reflect.Struct {
		return ErrInvalidInsertBatchValue
	}
	idName := s.builder.buildBatchValues(value)

	Sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + Sql)
	}
	return handlerValuesReturningBatch(s.conn, Sql, value, s.builder.ArgsAny, idName, s.ctx)
}

/*
State Update
*/
type stateUpdate struct {
	config  *Config
	conn    Connection
	addrMap map[uintptr]Field
	builder *Builder
	ctx     context.Context
	err     error
}

func createUpdateState(am map[uintptr]Field, conn Connection, c *Config, ctx context.Context, d Driver, e error) *stateUpdate {
	return &stateUpdate{addrMap: am, conn: conn, builder: CreateBuilder(d), config: c, ctx: ctx, err: e}
}

func (s *stateUpdate) Where(Brs ...wh.Operator) *stateUpdate {
	if s.err != nil {
		return s
	}
	s.err = helperWhere(s.builder, s.addrMap, Brs...)
	return s
}

func (s *stateUpdate) queryUpdate(Args []uintptr, addrMap map[uintptr]Field) *stateUpdate {
	if s.err == nil {
		s.builder.Args = Args
		s.builder.buildUpdate(addrMap)
	}
	return s
}

// Value updates the targets in the database.
//
// The value can be a pointer to struct or a struct value.
//
// Value returns the SQL generated and error as nil if update with success.
//
// # Example
//
//	// updates all rows with aStruct values
//	db.Update(db.Animal).Value(aStruct)
//
//	// updates single row using where
//	db.Update(db.Animal).Where(db.Equals(&db.Animal.Id, aStruct.Id)).Value(aStruct)
func (s *stateUpdate) Value(value any) error {
	if s.err != nil {
		return s.err
	}

	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return ErrInvalidUpdateValue
	}

	s.builder.buildSet(v)

	//generate query
	s.err = s.builder.buildSqlUpdate()
	if s.err != nil {
		return s.err
	}

	Sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + Sql)
	}
	return handlerValues(s.conn, Sql, s.builder.ArgsAny, s.ctx)
}

type stateDelete struct {
	addrMap map[uintptr]Field
	config  *Config
	conn    Connection
	builder *Builder
	ctx     context.Context
	err     error
}

func createDeleteState(am map[uintptr]Field, conn Connection, c *Config, ctx context.Context, d Driver, e error) *stateDelete {
	return &stateDelete{addrMap: am, conn: conn, builder: CreateBuilder(d), config: c, ctx: ctx, err: e}
}

func (s *stateDelete) queryDelete(Args []uintptr, addrMap map[uintptr]Field) *stateDelete {
	if s.err == nil {
		s.builder.Args = Args
		s.builder.buildDelete(addrMap)
	}
	return s
}

// Where from state delete executes the delete command in the database.
//
// # Example
//
//	// delete all animals
//	db.Delete(db.Animal).Where()
//
//	// delete matched animals
//	db.Delete(db.Animal).Where(wh.Equals(&db.Animal.Id, 23))
func (s *stateDelete) Where(Brs ...wh.Operator) error {
	if s.err != nil {
		return s.err
	}

	s.err = helperWhere(s.builder, s.addrMap, Brs...)
	if s.err != nil {
		return s.err
	}

	s.err = s.builder.buildSqlDelete()
	if s.err != nil {
		return s.err
	}

	Sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + Sql)
	}
	return handlerValues(s.conn, Sql, s.builder.ArgsAny, s.ctx)
}

func helperWhere(builder *Builder, addrMap map[uintptr]Field, Brs ...wh.Operator) error {
	for i := range Brs {
		switch br := Brs[i].(type) {
		case wh.Operation:
			if a := getArg(br.Arg, addrMap); a != nil {
				br.Arg = a.GetSelect()
				builder.Brs = append(builder.Brs, br)
				continue
			}
			return ErrInvalidWhere
		case wh.OperationArg:
			if a, b := getArg(br.Op.Arg, addrMap), getArg(br.Op.Value, addrMap); a != nil && b != nil {
				br.Op.Arg = a.GetSelect()
				br.Op.ValueFlag = b.GetSelect()
				builder.Brs = append(builder.Brs, br)
				continue
			}
			return ErrInvalidWhere
		case wh.OperationIs:
			if a := getArg(br.Arg, addrMap); a != nil {
				br.Arg = a.GetSelect()
				builder.Brs = append(builder.Brs, br)
				continue
			}
			return ErrInvalidWhere
		default:
			builder.Brs = append(builder.Brs, br)
		}
	}
	return nil
}
