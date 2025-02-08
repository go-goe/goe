package goe

import (
	"context"
	"log"
	"reflect"
	"slices"

	"github.com/olauro/goe/query"
)

type save[T any] struct {
	table    *T
	pks      []field
	pksValue []any
	includes []field
	update   *stateUpdate[T]
}

func Save[T any](table *T, tx ...*Tx) *save[T] {
	return SaveContext(context.Background(), table, tx...)
}

func SaveContext[T any](ctx context.Context, table *T, tx ...*Tx) *save[T] {
	save := &save[T]{}
	save.update = UpdateContext(ctx, table, tx...)

	if save.update.err != nil {
		return save
	}

	save.table = table

	return save
}

func (s *save[T]) Includes(args ...any) *save[T] {
	ptrArgs, err := getArgsUpdate(addrMap, args...)
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

	s.pks, s.pksValue, s.update.err = getPksField(addrMap, s.table, replace)
	return s
}

func (s *save[T]) Value(v T) error {
	if s.update.err != nil {
		return s.update.err
	}

	includes, pks, pksValue, err := getArgsSave(addrMap, s.table, v)
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
		if !slices.ContainsFunc(includes, func(f field) bool {
			//TODO: Add Id to compare
			return f.getSelect() == s.includes[i].getSelect()
		}) {
			includes = append(includes, s.includes[i])
		}
	}

	s.update.builder.fields = includes
	return s.update.Value(v)
}

type stateUpdate[T any] struct {
	config  *Config
	conn    Connection
	builder *builder
	ctx     context.Context
	err     error
}

// Update uses [context.Background] internally;
// to specify the context, use [DB.UpdateContext].
//
// # Example
func Update[T any](table *T, tx ...*Tx) *stateUpdate[T] {
	return UpdateContext(context.Background(), table, tx...)
}

// UpdateContext creates a update state for table
func UpdateContext[T any](ctx context.Context, table *T, tx ...*Tx) *stateUpdate[T] {
	f := getArg(table, addrMap)

	var state *stateUpdate[T]
	if f == nil {
		state = new(stateUpdate[T])
		state.err = ErrInvalidArg
		return state
	}

	db := f.getDb()

	if tx != nil {
		state = createUpdateState[T](tx[0].SqlTx, db.Config, ctx, db.Driver, nil)
	} else {
		state = createUpdateState[T](db.SqlDB, db.Config, ctx, db.Driver, nil)
	}

	return state
}

func (s *stateUpdate[T]) Includes(args ...any) *stateUpdate[T] {
	if s.err != nil {
		return s
	}

	fields, err := getArgsUpdate(addrMap, args...)

	if err != nil {
		s.err = err
		return s
	}

	s.builder.fields = append(s.builder.fields, fields...)
	return s
}

func (s *stateUpdate[T]) Where(brs ...query.Operator) *stateUpdate[T] {
	if s.err != nil {
		return s
	}
	s.err = helperWhere(s.builder, addrMap, brs...)
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

	s.err = s.builder.buildSqlUpdate(v)
	if s.err != nil {
		return s.err
	}

	sql := s.builder.sql.String()
	if s.config.LogQuery {
		log.Println("\n" + sql)
	}
	return handlerValues(s.conn, sql, s.builder.argsAny, s.ctx)
}

func getArgsUpdate(addrMap map[uintptr]field, args ...any) ([]field, error) {
	fields := make([]field, 0)
	var valueOf reflect.Value
	for i := range args {
		valueOf = reflect.ValueOf(args[i])

		if valueOf.Kind() != reflect.Pointer {
			return nil, ErrInvalidArg
		}

		valueOf = valueOf.Elem()

		if !valueOf.CanAddr() {
			return nil, ErrInvalidArg
		}

		addr := uintptr(valueOf.Addr().UnsafePointer())
		if addrMap[addr] != nil {
			fields = append(fields, addrMap[addr])
		}
	}
	if len(fields) == 0 {
		return nil, ErrInvalidArg
	}
	return fields, nil
}

func getArgsSave[T any](addrMap map[uintptr]field, table *T, value T) ([]field, []field, []any, error) {
	if table == nil {
		return nil, nil, nil, ErrInvalidArg
	}

	tableOf := reflect.ValueOf(table).Elem()

	args, pks, pksValue := make([]field, 0), make([]field, 0), make([]any, 0)

	if tableOf.Kind() != reflect.Struct {
		return nil, nil, nil, ErrInvalidArg
	}

	valueOf := reflect.ValueOf(value)
	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			if addrMap[addr] != nil {
				if addrMap[addr].isPrimaryKey() {
					pks = append(pks, addrMap[addr])
					pksValue = append(pksValue, valueOf.Field(i).Interface())
					continue
				}
				args = append(args, addrMap[addr])
			}
		}
	}

	if len(args) == 0 && len(pks) == 0 {
		return nil, nil, nil, ErrInvalidArg
	}
	return args, pks, pksValue, nil
}

func getPksField[T any](addrMap map[uintptr]field, table *T, value T) ([]field, []any, error) {
	pks, pksValue := make([]field, 0), make([]any, 0)

	tableOf := reflect.ValueOf(table).Elem()
	if tableOf.Kind() != reflect.Struct {
		return nil, nil, ErrInvalidArg
	}

	valueOf := reflect.ValueOf(value)
	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			if addrMap[addr] != nil {
				if addrMap[addr].isPrimaryKey() {
					pks = append(pks, addrMap[addr])
					pksValue = append(pksValue, valueOf.Field(i).Interface())
				}
			}
		}
	}

	if len(pks) == 0 {
		return nil, nil, ErrInvalidArg
	}
	return pks, pksValue, nil
}

func helperOperation(builder *builder, pks []field, pksValue []any) {
	builder.brs = append(builder.brs, query.Operation{
		Arg:      pks[0].getSelect(),
		Operator: "=",
		Value:    pksValue[0]})
	pkCount := 1
	for _, pk := range pks[1:] {
		builder.brs = append(builder.brs, query.And())
		builder.brs = append(builder.brs, query.Operation{
			Arg:      pk.getSelect(),
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
	return &stateUpdate[T]{conn: conn, builder: createBuilder(d), config: c, ctx: ctx, err: e}
}
