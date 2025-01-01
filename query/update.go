package query

import (
	"context"
	"log"
	"reflect"
	"slices"

	"github.com/olauro/goe"
	"github.com/olauro/goe/wh"
)

type save[T any] struct {
	table    *T
	pks      []uintptr
	pksValue []any
	includes []uintptr
	update   *stateUpdate[T]
}

func Save[T any](db *goe.DB, table *T) *save[T] {
	return SaveContext(context.Background(), db, table)
}

func SaveContext[T any](ctx context.Context, db *goe.DB, table *T) *save[T] {
	save := &save[T]{}
	save.update = UpdateContext(ctx, db, table)

	if table == nil {
		save.update.err = goe.ErrInvalidArg
		return save
	}

	save.table = table

	return save
}

func (s *save[T]) Includes(args ...any) *save[T] {
	ptrArgs, err := getArgsUpdate(s.update.addrMap, args...)
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

	s.pks, s.pksValue, s.update.err = getArgsReplace(s.update.addrMap, s.table, replace)
	return s
}

func (s *save[T]) Value(v T) error {
	if s.update.err != nil {
		return s.update.err
	}

	includes, pks, pksValue, err := getArgsSave(s.update.addrMap, s.table, v)
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
		return goe.ErrInvalidArg
	}

	s.update.builder.Brs = append(s.update.builder.Brs, wh.Operation{
		Arg:      s.update.addrMap[pks[0]].GetSelect(),
		Operator: "=",
		Value:    pksValue[0]})
	pkCount := 1
	for _, pk := range pks[1:] {
		s.update.builder.Brs = append(s.update.builder.Brs, wh.Logical{Operator: "AND"})
		s.update.builder.Brs = append(s.update.builder.Brs, wh.Operation{
			Arg:      s.update.addrMap[pk].GetSelect(),
			Operator: "=",
			Value:    pksValue[pkCount]})
		pkCount++
	}

	for i := range s.includes {
		if !slices.Contains(includes, s.includes[i]) {
			includes = append(includes, s.includes[i])
		}
	}
	s.update.queryUpdate(includes, s.update.addrMap)

	return s.update.Value(v)
}

type stateUpdate[T any] struct {
	config  *goe.Config
	conn    goe.Connection
	addrMap map[uintptr]goe.Field
	builder *goe.Builder
	ctx     context.Context
	err     error
}

func Update[T any](db *goe.DB, table *T) *stateUpdate[T] {
	return UpdateContext[T](context.Background(), db, table)
}

// UpdateContext creates a update state for table
func UpdateContext[T any](ctx context.Context, db *goe.DB, table *T) *stateUpdate[T] {
	s := createUpdateState[T](db.AddrMap, db.ConnPool, db.Config, ctx, db.Driver, nil)
	if table == nil {
		s.err = goe.ErrInvalidArg
	}
	return s
}

func (s *stateUpdate[T]) Includes(args ...any) *stateUpdate[T] {
	if s.err != nil {
		return s
	}

	ptrArgs, err := getArgsUpdate(s.addrMap, args...)

	if err != nil {
		s.err = err
		return s
	}

	return s.queryUpdate(ptrArgs, s.addrMap)
}

func (s *stateUpdate[T]) Where(brs ...wh.Operator) *stateUpdate[T] {
	if s.err != nil {
		return s
	}
	s.err = helperWhere(s.builder, s.addrMap, brs...)
	return s
}

func (s *stateUpdate[T]) Value(value T) error {
	if s.err != nil {
		return s.err
	}

	if s.builder.Args == nil {
		//TODO: Includes error
		return goe.ErrInvalidArg
	}

	v := reflect.ValueOf(value)

	s.builder.BuildSet(v)

	//generate query
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

func (s *stateUpdate[T]) queryUpdate(args []uintptr, addrMap map[uintptr]goe.Field) *stateUpdate[T] {
	if s.err == nil {
		s.builder.Args = append(s.builder.Args, args...)
		s.builder.BuildUpdate(addrMap)
	}
	return s
}

func getArgsUpdate(AddrMap map[uintptr]goe.Field, args ...any) ([]uintptr, error) {
	ptrArgs := make([]uintptr, 0)
	var valueOf reflect.Value
	for i := range args {
		valueOf = reflect.ValueOf(args[i])

		if valueOf.Kind() != reflect.Pointer {
			return nil, goe.ErrInvalidArg
		}

		valueOf = valueOf.Elem()
		addr := uintptr(valueOf.Addr().UnsafePointer())
		if AddrMap[addr] != nil {
			ptrArgs = append(ptrArgs, addr)
		}
	}
	if len(ptrArgs) == 0 {
		return nil, goe.ErrInvalidArg
	}
	return ptrArgs, nil
}

func getArgsSave[T any](AddrMap map[uintptr]goe.Field, table *T, value T) ([]uintptr, []uintptr, []any, error) {
	if table == nil {
		return nil, nil, nil, goe.ErrInvalidArg
	}

	tableOf := reflect.ValueOf(table).Elem()

	args, pks, pksValue := make([]uintptr, 0), make([]uintptr, 0), make([]any, 0)

	if tableOf.Kind() != reflect.Struct {
		return nil, nil, nil, goe.ErrInvalidArg
	}

	valueOf := reflect.ValueOf(value)
	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			//TODO: Update to Field
			if AddrMap[addr] != nil {
				if AddrMap[addr].IsPrimaryKey() {
					pks = append(pks, addr)
					pksValue = append(pksValue, valueOf.Field(i).Interface())
					continue
				}
				args = append(args, addr)
			}
		}
	}

	if len(args) == 0 && len(pks) == 0 {
		return nil, nil, nil, goe.ErrInvalidArg
	}
	return args, pks, pksValue, nil
}

func getArgsReplace[T any](AddrMap map[uintptr]goe.Field, table *T, value T) ([]uintptr, []any, error) {
	pks, pksValue := make([]uintptr, 0), make([]any, 0)

	tableOf := reflect.ValueOf(table).Elem()
	if tableOf.Kind() != reflect.Struct {
		return nil, nil, goe.ErrInvalidArg
	}

	valueOf := reflect.ValueOf(value)
	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			//TODO: Update to Field
			if AddrMap[addr] != nil {
				if AddrMap[addr].IsPrimaryKey() {
					pks = append(pks, addr)
					pksValue = append(pksValue, valueOf.Field(i).Interface())
					continue
				}
			}
		}
	}

	if len(pks) == 0 {
		return nil, nil, goe.ErrInvalidArg
	}
	return pks, pksValue, nil
}

func createUpdateState[T any](
	am map[uintptr]goe.Field,
	conn goe.Connection, c *goe.Config,
	ctx context.Context,
	d goe.Driver,
	e error) *stateUpdate[T] {
	return &stateUpdate[T]{addrMap: am, conn: conn, builder: goe.CreateBuilder(d), config: c, ctx: ctx, err: e}
}
