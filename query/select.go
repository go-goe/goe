package query

import (
	"context"
	"fmt"
	"iter"
	"log"
	"reflect"

	"github.com/olauro/goe"
	"github.com/olauro/goe/jn"
	"github.com/olauro/goe/wh"
)

type stateSelect[T any] struct {
	config  *goe.Config
	conn    goe.Connection
	addrMap map[uintptr]goe.Field
	builder *goe.Builder
	ctx     context.Context
	err     error
}

func Find[T any, P any](db *goe.DB, t *T, pk *P, v P) (T, error) {
	return FindContext(context.Background(), db, t, pk, v)
}

func FindContext[T any, P any](ctx context.Context, db *goe.DB, t *T, pk *P, v P) (T, error) {
	var (
		row T
		err error
	)
	for row, err = range SelectContext(ctx, db, t).From(t).
		Where(wh.Equals(pk, v)).Rows() {
		return row, err
	}
	return row, goe.ErrNotFound
}

func Select[T any](db *goe.DB, t *T) *stateSelect[T] {
	return SelectContext(context.Background(), db, t)
}

func SelectContext[T any](ctx context.Context, db *goe.DB, t *T) *stateSelect[T] {
	ts, err := getArgsSelect(db.AddrMap, t)

	var state *stateSelect[T]
	if err != nil {
		state = createSelectState[T](nil, db.Config, ctx, nil, err)
		return state.querySelect(nil)
	}

	state = createSelectState[T](db.ConnPool, db.Config, ctx, db.Driver, err)

	state.addrMap = db.AddrMap
	return state.querySelect(ts)
}

// Where creates a where SQL using the operations
func (s *stateSelect[T]) Where(brs ...wh.Operator) *stateSelect[T] {
	if s.err != nil {
		return s
	}
	s.err = helperWhere(s.builder, s.addrMap, brs...)
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
func (s *stateSelect[T]) Take(i uint) *stateSelect[T] {
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
func (s *stateSelect[T]) Skip(i uint) *stateSelect[T] {
	s.builder.Offset = i
	return s
}

// Page returns page p with i elements
//
// # Example
//
//	// returns first 20 elements
//	db.Select(db.Habitat).Page(1, 20).Scan(&h)
func (s *stateSelect[T]) Page(p uint, i uint) *stateSelect[T] {
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
func (s *stateSelect[T]) OrderByAsc(arg any) *stateSelect[T] {
	Field := getArg(arg, s.addrMap)
	if Field == nil {
		s.err = goe.ErrInvalidOrderBy
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
func (s *stateSelect[T]) OrderByDesc(arg any) *stateSelect[T] {
	Field := getArg(arg, s.addrMap)
	if Field == nil {
		s.err = goe.ErrInvalidOrderBy
		return s
	}
	s.builder.OrderBy = fmt.Sprintf("\nORDER BY %v DESC", Field.GetSelect())
	return s
}

// TODO: Add Doc
func (s *stateSelect[T]) From(tables ...any) *stateSelect[T] {
	s.builder.Tables = make([]string, len(tables))
	args, err := getArgsTables(s.addrMap, s.builder.Tables, tables...)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.Froms = args
	return s
}

// TODO: Add Doc
func (s *stateSelect[T]) Joins(joins ...jn.Joins) *stateSelect[T] {
	if s.err != nil {
		return s
	}

	for _, j := range joins {
		args, err := getArgsIn(s.addrMap, j.FirstArg(), j.SecondArg())
		if err != nil {
			s.err = err
			return s
		}
		s.builder.BuildSelectJoins(s.addrMap, j.Join(), args)
	}
	return s
}

// TODO: Add doc
func (s *stateSelect[T]) Rows() iter.Seq2[T, error] {
	if s.err != nil {
		var v T
		return func(yield func(T, error) bool) {
			yield(v, s.err)
		}
	}

	//generate query
	s.err = s.builder.BuildSqlSelect()
	if s.err != nil {
		var v T
		return func(yield func(T, error) bool) {
			yield(v, s.err)
		}
	}

	Sql := s.builder.Sql.String()
	if s.config.LogQuery {
		log.Println("\n" + Sql)
	}

	return handlerResult[T](s.conn, Sql, s.builder.ArgsAny, s.builder.StructColumns, s.ctx)
}

func SafeGet[T any](v *T) T {
	if v == nil {
		return reflect.New(reflect.TypeOf(v).Elem()).Elem().Interface().(T)
	}
	return *v
}

func createSelectState[T any](conn goe.Connection, c *goe.Config, ctx context.Context, d goe.Driver, e error) *stateSelect[T] {
	return &stateSelect[T]{conn: conn, builder: goe.CreateBuilder(d), config: c, ctx: ctx, err: e}
}

func (s *stateSelect[T]) querySelect(args []uintptr) *stateSelect[T] {
	if s.err == nil {
		s.builder.Args = args
		s.builder.BuildSelect(s.addrMap)
	}
	return s
}

func getArgsSelect(AddrMap map[uintptr]goe.Field, arg any) ([]uintptr, error) {
	uintArgs := make([]uintptr, 0)

	if reflect.ValueOf(arg).Kind() != reflect.Pointer {
		return nil, goe.ErrInvalidArg
	}

	valueOf := reflect.ValueOf(arg).Elem()

	if valueOf.Kind() != reflect.Struct {
		return nil, goe.ErrInvalidArg
	}

	var fieldOf reflect.Value
	for i := 0; i < valueOf.NumField(); i++ {
		fieldOf = valueOf.Field(i)
		if fieldOf.Kind() == reflect.Slice && fieldOf.Type().Elem().Kind() == reflect.Struct {
			continue
		}
		addr := uintptr(fieldOf.Addr().UnsafePointer())
		if AddrMap[addr] != nil {
			uintArgs = append(uintArgs, addr)
			continue
		}
		//get args from anonymous struct
		return getArgsSelectAno(AddrMap, valueOf)
	}

	return uintArgs, nil
}

func getArgsSelectAno(AddrMap map[uintptr]goe.Field, valueOf reflect.Value) ([]uintptr, error) {
	uintArgs := make([]uintptr, 0)
	var fieldOf reflect.Value
	for i := 0; i < valueOf.NumField(); i++ {
		fieldOf = valueOf.Field(i).Elem()
		addr := uintptr(fieldOf.Addr().UnsafePointer())
		if AddrMap[addr] != nil {
			uintArgs = append(uintArgs, addr)
			continue
		}
		return nil, goe.ErrInvalidArg
	}
	if len(uintArgs) == 0 {
		return nil, goe.ErrInvalidArg
	}
	return uintArgs, nil
}

func getArgsIn(AddrMap map[uintptr]goe.Field, args ...any) ([]uintptr, error) {
	stringArgs := make([]uintptr, 2)
	var ptr uintptr
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			ptr = uintptr(valueOf.Addr().UnsafePointer())
			if AddrMap[ptr] != nil {
				stringArgs[i] = ptr
			}
		} else {
			return nil, goe.ErrInvalidArg
		}
	}

	if stringArgs[0] == 0 || stringArgs[1] == 0 {
		return nil, goe.ErrInvalidArg
	}
	return stringArgs, nil
}

func getArgsTables(AddrMap map[uintptr]goe.Field, tables []string, args ...any) ([]byte, error) {
	from := make([]byte, 0)
	var ptr uintptr
	var i int
	if reflect.ValueOf(args[0]).Kind() == reflect.Ptr {
		valueOf := reflect.ValueOf(args[0]).Elem()
		ptr = uintptr(valueOf.Addr().UnsafePointer())
		if AddrMap[ptr] == nil {
			//TODO: add ErrInvalidTable
			return nil, goe.ErrInvalidArg
		}
		tables[i] = string(AddrMap[ptr].Table())
		i++
		from = append(from, AddrMap[ptr].Table()...)
	} else {
		return nil, goe.ErrInvalidArg
	}
	for _, a := range args[1:] {
		if reflect.ValueOf(a).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(a).Elem()
			ptr = uintptr(valueOf.Addr().UnsafePointer())
			if AddrMap[ptr] == nil {
				//TODO: add ErrInvalidTable
				return nil, goe.ErrInvalidArg
			}
			tables[i] = string(AddrMap[ptr].Table())
			i++
			from = append(from, ',')
			from = append(from, AddrMap[ptr].Table()...)
		} else {
			return nil, goe.ErrInvalidArg
		}
	}

	return from, nil
}

func getArg(arg any, AddrMap map[uintptr]goe.Field) goe.Field {
	v := reflect.ValueOf(arg)
	if v.Kind() != reflect.Pointer {
		return nil
	}

	addr := uintptr(v.UnsafePointer())
	if AddrMap[addr] != nil {
		return AddrMap[addr]
	}
	return nil
}

func helperWhere(builder *goe.Builder, addrMap map[uintptr]goe.Field, brs ...wh.Operator) error {
	for i := range brs {
		switch br := brs[i].(type) {
		case wh.Operation:
			if a := getArg(br.Arg, addrMap); a != nil {
				br.Arg = a.GetSelect()
				builder.Brs = append(builder.Brs, br)
				continue
			}
			return goe.ErrInvalidWhere
		case wh.OperationArg:
			if a, b := getArg(br.Op.Arg, addrMap), getArg(br.Op.Value, addrMap); a != nil && b != nil {
				br.Op.Arg = a.GetSelect()
				br.Op.ValueFlag = b.GetSelect()
				builder.Brs = append(builder.Brs, br)
				continue
			}
			return goe.ErrInvalidWhere
		case wh.OperationIs:
			if a := getArg(br.Arg, addrMap); a != nil {
				br.Arg = a.GetSelect()
				builder.Brs = append(builder.Brs, br)
				continue
			}
			return goe.ErrInvalidWhere
		default:
			builder.Brs = append(builder.Brs, br)
		}
	}
	return nil
}
