package goe

import (
	"context"
	"log"
	"reflect"
	"slices"

	"github.com/olauro/goe/wh"
)

type save[T any] struct {
	table    *T
	pks      []Field
	pksValue []any
	includes []Field
	update   *stateUpdate[T]
}

func Save[T any](table *T) *save[T] {
	return SaveContext(context.Background(), table)
}

func SaveContext[T any](ctx context.Context, table *T) *save[T] {
	save := &save[T]{}
	save.update = UpdateContext(ctx, table)

	if save.update.err != nil {
		return save
	}

	save.table = table

	return save
}

func (s *save[T]) Includes(args ...any) *save[T] {
	ptrArgs, err := getArgsUpdate(AddrMap, args...)
	if err != nil {
		s.update.err = err
		return s
	}

	s.includes = append(s.includes, ptrArgs...)
	return s
}

func (s *save[T]) Replace(replace T) *save[T] {
	if s.update.err != nil {
		return s
	}

	s.pks, s.pksValue, s.update.err = getArgsPks(AddrMap, s.table, replace)
	return s
}

func (s *save[T]) Value(v T) error {
	if s.update.err != nil {
		return s.update.err
	}

	includes, pks, pksValue, err := getArgsSave(AddrMap, s.table, v)
	if err != nil {
		return err
	}

	if s.pks != nil {
		for i := range pks {
			if pksValue[i] != s.pksValue[i] {
				includes = append(includes, pks[i])
			}
		}
		pks = s.pks
		pksValue = s.pksValue
	}

	if len(includes) == 0 {
		//TODO: error inclues empty
		return ErrInvalidArg
	}

	helperOperation(s.update.builder, pks, pksValue)

	for i := range s.includes {
		if !slices.ContainsFunc(includes, func(f Field) bool {
			//TODO: Add Id to compare
			return f.GetSelect() == s.includes[i].GetSelect()
		}) {
			includes = append(includes, s.includes[i])
		}
	}

	s.update.builder.Fields = includes
	return s.update.Value(v)
}

type stateUpdate[T any] struct {
	config  *Config
	conn    Connection
	builder *Builder
	ctx     context.Context
	err     error
}

// Update uses [context.Background] internally;
// to specify the context, use [DB.UpdateContext].
//
// # Example
func Update[T any](table *T) *stateUpdate[T] {
	return UpdateContext[T](context.Background(), table)
}

// UpdateContext creates a update state for table
func UpdateContext[T any](ctx context.Context, table *T) *stateUpdate[T] {
	f := getArg(table, AddrMap)
	var s *stateUpdate[T]
	if f == nil {
		s = new(stateUpdate[T])
		s.err = ErrInvalidArg
		return s
	}
	db := f.GetDb()
	s = createUpdateState[T](db.ConnPool, db.Config, ctx, db.Driver, nil)

	return s
}

func (s *stateUpdate[T]) Includes(args ...any) *stateUpdate[T] {
	if s.err != nil {
		return s
	}

	fields, err := getArgsUpdate(AddrMap, args...)

	if err != nil {
		s.err = err
		return s
	}

	s.builder.Fields = append(s.builder.Fields, fields...)
	return s
}

func (s *stateUpdate[T]) Where(brs ...wh.Operator) *stateUpdate[T] {
	if s.err != nil {
		return s
	}
	s.err = helperWhere(s.builder, AddrMap, brs...)
	return s
}

func (s *stateUpdate[T]) Value(value T) error {
	if s.err != nil {
		return s.err
	}

	if s.conn == nil {
		//TODO: Includes error
		return ErrInvalidArg
	}

	v := reflect.ValueOf(value)

	s.builder.BuildUpdate()
	s.builder.BuildSet(v)
	s.err = s.builder.BuildSqlUpdate()
	if s.err != nil {
		return s.err
	}

	sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + sql)
	}
	return handlerValues(s.conn, sql, s.builder.ArgsAny, s.ctx)
}

func getArgsUpdate(AddrMap map[uintptr]Field, args ...any) ([]Field, error) {
	fields := make([]Field, 0)
	var valueOf reflect.Value
	for i := range args {
		valueOf = reflect.ValueOf(args[i])

		if valueOf.Kind() != reflect.Pointer {
			return nil, ErrInvalidArg
		}

		valueOf = valueOf.Elem()
		addr := uintptr(valueOf.Addr().UnsafePointer())
		if AddrMap[addr] != nil {
			fields = append(fields, AddrMap[addr])
		}
	}
	if len(fields) == 0 {
		return nil, ErrInvalidArg
	}
	return fields, nil
}

func getArgsSave[T any](AddrMap map[uintptr]Field, table *T, value T) ([]Field, []Field, []any, error) {
	if table == nil {
		return nil, nil, nil, ErrInvalidArg
	}

	tableOf := reflect.ValueOf(table).Elem()

	args, pks, pksValue := make([]Field, 0), make([]Field, 0), make([]any, 0)

	if tableOf.Kind() != reflect.Struct {
		return nil, nil, nil, ErrInvalidArg
	}

	valueOf := reflect.ValueOf(value)
	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			if AddrMap[addr] != nil {
				if AddrMap[addr].IsPrimaryKey() {
					pks = append(pks, AddrMap[addr])
					pksValue = append(pksValue, valueOf.Field(i).Interface())
					continue
				}
				args = append(args, AddrMap[addr])
			}
		}
	}

	if len(args) == 0 && len(pks) == 0 {
		return nil, nil, nil, ErrInvalidArg
	}
	return args, pks, pksValue, nil
}

func getArgsPks[T any](AddrMap map[uintptr]Field, table *T, value T) ([]Field, []any, error) {
	pks, pksValue := make([]Field, 0), make([]any, 0)

	tableOf := reflect.ValueOf(table).Elem()
	if tableOf.Kind() != reflect.Struct {
		return nil, nil, ErrInvalidArg
	}

	valueOf := reflect.ValueOf(value)
	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			//TODO: Update to Field
			if AddrMap[addr] != nil {
				if AddrMap[addr].IsPrimaryKey() {
					pks = append(pks, AddrMap[addr])
					pksValue = append(pksValue, valueOf.Field(i).Interface())
					continue
				}
			}
		}
	}

	if len(pks) == 0 {
		return nil, nil, ErrInvalidArg
	}
	return pks, pksValue, nil
}

func helperOperation(builder *Builder, pks []Field, pksValue []any) {
	builder.Brs = append(builder.Brs, wh.Operation{
		Arg:      pks[0].GetSelect(),
		Operator: "=",
		Value:    pksValue[0]})
	pkCount := 1
	for _, pk := range pks[1:] {
		builder.Brs = append(builder.Brs, wh.Logical{Operator: "AND"})
		builder.Brs = append(builder.Brs, wh.Operation{
			Arg:      pk.GetSelect(),
			Operator: "=",
			Value:    pksValue[pkCount]})
		pkCount++
	}
}

func createUpdateState[T any](
	conn Connection, c *Config,
	ctx context.Context,
	d Driver,
	e error) *stateUpdate[T] {
	return &stateUpdate[T]{conn: conn, builder: CreateBuilder(d), config: c, ctx: ctx, err: e}
}
