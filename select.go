package goe

import (
	"context"
	"fmt"
	"iter"
	"log"
	"reflect"
	"strings"

	"github.com/olauro/goe/query"
)

type stateSelect[T any] struct {
	config  *Config
	conn    Connection
	builder *builder
	ctx     context.Context
	err     error
}

func Find[T any](t *T, v T, tx ...*Tx) (*T, error) {
	return FindContext(context.Background(), t, v, tx...)
}

func FindContext[T any](ctx context.Context, t *T, v T, tx ...*Tx) (*T, error) {
	pks, pksValue, err := getPksField(addrMap, t, v)

	if err != nil {
		return nil, err
	}

	s := SelectContext(ctx, t, tx...).From(t)
	helperOperation(s.builder, pks, pksValue)

	for row, err := range s.Rows() {
		if err != nil {
			return nil, err
		}
		return &row, nil
	}

	return nil, ErrNotFound
}

// Select uses [context.Background] internally;
// to specify the context, use [query.SelectContext].
//
// # Example
func Select[T any](t *T, tx ...*Tx) *stateSelect[T] {
	return SelectContext(context.Background(), t, tx...)
}

func SelectContext[T any](ctx context.Context, t *T, tx ...*Tx) *stateSelect[T] {
	fields, err := getArgsSelect(addrMap, t)

	var state *stateSelect[T]
	if err != nil {
		state = new(stateSelect[T])
		state.err = err
		return state
	}

	db := fields[0].getDb()

	if tx != nil {
		state = createSelectState[T](tx[0].SqlTx, db.Config, ctx, db.Driver, nil)
	} else {
		state = createSelectState[T](db.SqlDB, db.Config, ctx, db.Driver, nil)
	}

	state.builder.fieldsSelect = fields
	return state
}

// Where creates a where SQL using the operations
func (s *stateSelect[T]) Where(brs ...query.Operator) *stateSelect[T] {
	if s.err != nil {
		return s
	}
	s.err = helperWhere(s.builder, addrMap, brs...)
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
	s.builder.limit = i
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
	s.builder.offset = i
	return s
}

// Page returns page p with i elements
//
// # Example
//
//	// returns first 20 elements
//	db.Select(db.Habitat).Page(1, 20).Scan(&h)
func (s *stateSelect[T]) Page(p uint, i uint) *stateSelect[T] {
	s.builder.offset = i * (p - 1)
	s.builder.limit = i
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
	Field := getArg(arg, addrMap)
	if Field == nil {
		s.err = ErrInvalidOrderBy
		return s
	}
	s.builder.orderBy = fmt.Sprintf("\nORDER BY %v ASC", Field.getSelect())
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
	Field := getArg(arg, addrMap)
	if Field == nil {
		s.err = ErrInvalidOrderBy
		return s
	}
	s.builder.orderBy = fmt.Sprintf("\nORDER BY %v DESC", Field.getSelect())
	return s
}

// TODO: Add Doc
func (s *stateSelect[T]) From(tables ...any) *stateSelect[T] {
	if s.err != nil {
		return s
	}

	s.builder.tables = make([]uint, len(tables))
	args, err := getArgsTables(addrMap, s.builder.tables, tables...)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.froms = args
	return s
}

// TODO: Add Doc
func (s *stateSelect[T]) Joins(joins ...query.Joins) *stateSelect[T] {
	if s.err != nil {
		return s
	}

	for _, j := range joins {
		fields, err := getArgsJoin(addrMap, j.FirstArg(), j.SecondArg())
		if err != nil {
			s.err = err
			return s
		}
		s.builder.buildSelectJoins(j.Join(), fields)
	}
	return s
}

func (s *stateSelect[T]) RowsAsSlice() ([]T, error) {
	rows := make([]T, 0, s.builder.limit)
	for row, err := range s.Rows() {
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// TODO: Add doc
func (s *stateSelect[T]) Rows() iter.Seq2[T, error] {
	if s.err != nil {
		var v T
		return func(yield func(T, error) bool) {
			yield(v, s.err)
		}
	}

	s.err = s.builder.buildSqlSelect()
	if s.err != nil {
		var v T
		return func(yield func(T, error) bool) {
			yield(v, s.err)
		}
	}

	sql := s.builder.sql.String()
	if s.config.LogQuery {
		log.Println("\n" + sql)
	}

	return handlerResult[T](s.conn, sql, s.builder.argsAny, len(s.builder.fieldsSelect), s.ctx)
}

func SafeGet[T any](v *T) T {
	if v == nil {
		return reflect.New(reflect.TypeOf(v).Elem()).Elem().Interface().(T)
	}
	return *v
}

func createSelectState[T any](conn Connection, c *Config, ctx context.Context, d Driver, e error) *stateSelect[T] {
	return &stateSelect[T]{conn: conn, builder: createBuilder(d), config: c, ctx: ctx, err: e}
}

type list[T any] struct {
	table   *T
	sSelect *stateSelect[T]
	err     error
}

func List[T any](t *T, tx ...*Tx) *list[T] {
	return ListContext(context.Background(), t, tx...)
}

func ListContext[T any](ctx context.Context, t *T, tx ...*Tx) *list[T] {
	return &list[T]{sSelect: SelectContext(ctx, t, tx...).From(t), table: t}
}

func (l *list[T]) Page(page, size uint) *list[T] {
	l.sSelect.Page(page, size)
	return l
}

func (l *list[T]) OrderByAsc(a any) *list[T] {
	l.sSelect.OrderByAsc(a)
	return l
}

func (l *list[T]) OrderByDesc(a any) *list[T] {
	l.sSelect.OrderByDesc(a)
	return l
}

func (l *list[T]) Filter(v T) *list[T] {
	fields, fieldsValue, err := getNonZeroFields(addrMap, l.table, v)

	if err != nil {
		l.err = err
		return l
	}
	helperNonZeroOperation(l.sSelect.builder, fields, fieldsValue)
	return l
}

func (l *list[T]) AsSlice() ([]T, error) {
	if l.err != nil {
		return nil, l.err
	}
	return l.sSelect.RowsAsSlice()
}

func getNonZeroFields[T any](addrMap map[uintptr]field, table *T, value T) ([]field, []any, error) {
	fields, fieldsValue := make([]field, 0), make([]any, 0)

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
				fields = append(fields, addrMap[addr])
				fieldsValue = append(fieldsValue, valueOf.Field(i).Interface())
			}
		}
	}

	if len(fields) == 0 {
		return nil, nil, ErrInvalidArg
	}
	return fields, fieldsValue, nil
}

func helperNonZeroOperation(builder *builder, fields []field, fieldsValue []any) {
	builder.brs = append(builder.brs, equalsOrLike(fields[0].getSelect(), fieldsValue[0]))
	for i := 1; i < len(fields); i++ {
		builder.brs = append(builder.brs, query.Or())
		builder.brs = append(builder.brs, equalsOrLike(fields[i].getSelect(), fieldsValue[i]))
	}
}

func equalsOrLike(s string, a any) query.Operation {
	v, ok := a.(string)

	if !ok {
		return query.Operation{
			Arg:      s,
			Operator: "=",
			Value:    a}
	}

	if strings.Contains(v, "%") {
		return query.Operation{
			Arg:      fmt.Sprintf("UPPER(%v)", s),
			Operator: "LIKE",
			Value:    strings.ToUpper(v)}
	}

	return query.Operation{
		Arg:      fmt.Sprintf("UPPER(%v)", s),
		Operator: "=",
		Value:    strings.ToUpper(v)}
}

func getArgsSelect(addrMap map[uintptr]field, arg any) ([]fieldSelect, error) {
	fields := make([]fieldSelect, 0)

	if reflect.ValueOf(arg).Kind() != reflect.Pointer {
		return nil, ErrInvalidArg
	}

	valueOf := reflect.ValueOf(arg).Elem()

	if valueOf.Kind() != reflect.Struct {
		return nil, ErrInvalidArg
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
			continue
		}
		//get args from anonymous struct
		return getArgsSelectAno(addrMap, valueOf)
	}

	if len(fields) == 0 {
		return nil, ErrInvalidArg
	}

	return fields, nil
}

func getArgsSelectAno(addrMap map[uintptr]field, valueOf reflect.Value) ([]fieldSelect, error) {
	fields := make([]fieldSelect, 0)
	var fieldOf reflect.Value
	for i := 0; i < valueOf.NumField(); i++ {
		fieldOf = valueOf.Field(i).Elem()
		addr := uintptr(fieldOf.Addr().UnsafePointer())
		if addrMap[addr] != nil {
			fields = append(fields, addrMap[addr])
			continue
		}
		// check if is aggregate
		if fieldOf.Kind() == reflect.Struct {
			addr := uintptr(fieldOf.Field(0).Elem().UnsafePointer())
			if addrMap[addr] != nil {
				fields = append(fields, createAggregate(addrMap[addr], fieldOf.Interface()))
				continue
			}
		}
		return nil, ErrInvalidArg
	}
	if len(fields) == 0 {
		return nil, ErrInvalidArg
	}
	return fields, nil
}

func createAggregate(field field, a any) fieldSelect {
	switch ag := a.(type) {
	case query.Count:
		return &aggregate{selectName: ag.Aggregate(field.getSelect()), db: field.getDb()}
	}

	return nil
}

func getArgsJoin(addrMap map[uintptr]field, args ...any) ([]field, error) {
	fields := make([]field, 2)
	var ptr uintptr
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			ptr = uintptr(valueOf.Addr().UnsafePointer())
			if addrMap[ptr] != nil {
				fields[i] = addrMap[ptr]
			}
		} else {
			return nil, ErrInvalidArg
		}
	}

	if fields[0] == nil || fields[1] == nil {
		return nil, ErrInvalidArg
	}
	return fields, nil
}

func getArgsTables(addrMap map[uintptr]field, tables []uint, args ...any) ([]byte, error) {
	if reflect.ValueOf(args[0]).Kind() != reflect.Pointer {
		//TODO: add ErrInvalidTable
		return nil, ErrInvalidArg
	}

	from := make([]byte, 0)
	var ptr uintptr
	var i int

	valueOf := reflect.ValueOf(args[0]).Elem()
	ptr = uintptr(valueOf.Addr().UnsafePointer())
	if addrMap[ptr] == nil {
		//TODO: add ErrInvalidTable
		return nil, ErrInvalidArg
	}
	tables[i] = addrMap[ptr].getTableId()
	i++
	from = append(from, addrMap[ptr].table()...)

	for _, a := range args[1:] {
		if reflect.ValueOf(a).Kind() != reflect.Pointer {
			//TODO: add ErrInvalidTable
			return nil, ErrInvalidArg
		}

		valueOf = reflect.ValueOf(a).Elem()
		ptr = uintptr(valueOf.Addr().UnsafePointer())
		if addrMap[ptr] == nil {
			//TODO: add ErrInvalidTable
			return nil, ErrInvalidArg
		}
		tables[i] = addrMap[ptr].getTableId()
		i++
		from = append(from, ',')
		from = append(from, addrMap[ptr].table()...)
	}

	return from, nil
}

func getArg(arg any, addrMap map[uintptr]field) field {
	v := reflect.ValueOf(arg)
	if v.Kind() != reflect.Pointer {
		return nil
	}

	addr := uintptr(v.UnsafePointer())
	if addrMap[addr] != nil {
		return addrMap[addr]
	}
	return nil
}

func helperWhere(builder *builder, addrMap map[uintptr]field, brs ...query.Operator) error {
	for i := range brs {
		switch br := brs[i].(type) {
		case query.Operation:
			if a := getArg(br.Arg, addrMap); a != nil {
				br.Arg = a.getSelect()
				builder.brs = append(builder.brs, br)
				continue
			}
			return ErrInvalidWhere
		case query.OperationArg:
			if a, b := getArg(br.Op.Arg, addrMap), getArg(br.Op.Value, addrMap); a != nil && b != nil {
				br.Op.Arg = a.getSelect()
				br.Op.ValueFlag = b.getSelect()
				builder.brs = append(builder.brs, br)
				continue
			}
			return ErrInvalidWhere
		case query.OperationIs:
			if a := getArg(br.Arg, addrMap); a != nil {
				br.Arg = a.getSelect()
				builder.brs = append(builder.brs, br)
				continue
			}
			return ErrInvalidWhere
		default:
			builder.brs = append(builder.brs, br)
		}
	}
	return nil
}
