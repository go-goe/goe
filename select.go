package goe

import (
	"context"
	"iter"
	"math"
	"reflect"
	"strings"

	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/model"
	"github.com/olauro/goe/query"
	"github.com/olauro/goe/query/aggregate"
	"github.com/olauro/goe/query/function"
	"github.com/olauro/goe/query/where"
)

type stateSelect[T any] struct {
	conn    Connection
	builder builder
	tables  []any
	ctx     context.Context
	err     error
}

func Find[T any](t *T, v T, tx ...Transaction) (*T, error) {
	return FindContext(context.Background(), t, v, tx...)
}

func FindContext[T any](ctx context.Context, table *T, value T, tx ...Transaction) (*T, error) {
	pks, valuesPks, err := getArgsPks(addrMap, table, value)
	if err != nil {
		return nil, err
	}

	s := SelectContext(ctx, table, tx...).From(table)

	s.Where(where.Equals(&pks[0], valuesPks[0]))
	for i := 1; i < len(pks); i++ {
		s.Where(where.And())
		s.Where(where.Equals(&pks[i], valuesPks[i]))
	}

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
func Select[T any](t *T, tx ...Transaction) *stateSelect[T] {
	return SelectContext(context.Background(), t, tx...)
}

func SelectContext[T any](ctx context.Context, t *T, tx ...Transaction) *stateSelect[T] {
	fields, err := getArgsSelect(addrMap, t)

	var state *stateSelect[T]
	if err != nil {
		state = new(stateSelect[T])
		state.err = err
		return state
	}

	db := fields[0].getDb()

	if tx != nil {
		state = createSelectState[T](tx[0], ctx)
	} else {
		state = createSelectState[T](db.Driver.NewConnection(), ctx)
	}

	state.builder.fieldsSelect = fields
	return state
}

// Where creates a where SQL using the operations
func (s *stateSelect[T]) Where(brs ...query.Operation) *stateSelect[T] {
	if s.err != nil {
		return s
	}
	s.err = helperWhere(&s.builder, addrMap, brs...)
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
	s.builder.query.Limit = i
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
	s.builder.query.Offset = i
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
	field := getArg(arg, addrMap, nil)
	if field == nil {
		s.err = ErrInvalidOrderBy
		return s
	}
	s.builder.query.OrderBy = &model.OrderBy{Attribute: model.Attribute{Name: field.getAttributeName(), Table: field.table()}}
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
	field := getArg(arg, addrMap, nil)
	if field == nil {
		s.err = ErrInvalidOrderBy
		return s
	}
	s.builder.query.OrderBy = &model.OrderBy{
		Attribute: model.Attribute{Name: field.getAttributeName(), Table: field.table()},
		Desc:      true}
	return s
}

// TODO: Add Doc
func (s *stateSelect[T]) From(tables ...any) *stateSelect[T] {
	if s.err != nil {
		return s
	}

	s.builder.tables = make([]int, len(tables))
	err := getArgsTables(&s.builder, addrMap, s.builder.tables, tables...)
	if err != nil {
		s.err = err
		return s
	}
	//TODO add query tables
	s.tables = tables
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

func (s *stateSelect[T]) AsSlice() ([]T, error) {
	rows := make([]T, 0, s.builder.query.Limit)
	for row, err := range s.Rows() {
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (s *stateSelect[T]) AsQuery() (*model.Query, error) {
	if s.err != nil {
		return nil, s.err
	}

	return &s.builder.query, s.builder.buildSqlSelect()
}

type Pagination[T any] struct {
	TotalValues int64
	TotalPages  uint

	PageValues int
	PageSize   uint

	CurrentPage     uint
	HasPreviousPage bool
	PreviousPage    uint
	HasNextPage     bool
	NextPage        uint

	StartIndex uint
	EndIndex   uint
	Values     []T
}

func (s *stateSelect[T]) AsPagination(page, size uint) (*Pagination[T], error) {
	if s.err != nil {
		return nil, s.err
	}

	if size == 0 {
		// zero results for size equals 0
		return new(Pagination[T]), nil
	}

	var err error
	if page == 0 {
		// page 0 equals page 1
		page = 1
	}

	stateCount := Select(&struct {
		*query.Count
	}{
		Count: aggregate.Count(s.tables[0]),
	})

	// copy joins
	stateCount.builder.joins = s.builder.joins
	stateCount.builder.joinsArgs = s.builder.joinsArgs
	stateCount.builder.tables = make([]int, len(s.builder.tables))
	copy(stateCount.builder.tables, s.builder.tables)
	stateCount.builder.query.Tables = s.builder.query.Tables

	// copy wheres
	stateCount.builder.brs = s.builder.brs

	var count int64
	for row, err := range stateCount.Rows() {
		if err != nil {
			return nil, err
		}
		count = row.Value
		break
	}

	s.builder.query.Offset = size * (page - 1)
	s.builder.query.Limit = size

	p := new(Pagination[T])

	p.Values, err = s.AsSlice()
	if err != nil {
		return nil, err
	}

	p.TotalValues = count

	p.TotalPages = uint(math.Ceil(float64(count) / float64(size)))
	p.CurrentPage = page

	if page == p.TotalPages {
		p.NextPage = page
	} else {
		p.NextPage = page + 1
		p.HasNextPage = true
	}

	if page == 1 {
		p.PreviousPage = page
	} else {
		p.PreviousPage = page - 1
		p.HasPreviousPage = true
	}

	p.PageSize = size
	p.PageValues = len(p.Values)

	p.StartIndex = (page-1)*size + 1

	if !p.HasNextPage {
		p.EndIndex = uint(p.TotalValues)
	} else {
		p.EndIndex = size * page
	}

	return p, nil
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

	return handlerResult[T](s.conn, s.builder.query, len(s.builder.fieldsSelect), s.ctx)
}

func SafeGet[T any](v *T) T {
	if v == nil {
		return reflect.New(reflect.TypeOf(v).Elem()).Elem().Interface().(T)
	}
	return *v
}

func createSelectState[T any](conn Connection, ctx context.Context) *stateSelect[T] {
	return &stateSelect[T]{conn: conn, builder: createBuilder(enum.SelectQuery), ctx: ctx}
}

type list[T any] struct {
	table   *T
	sSelect *stateSelect[T]
	err     error
}

func List[T any](t *T, tx ...Transaction) *list[T] {
	return ListContext(context.Background(), t, tx...)
}

func ListContext[T any](ctx context.Context, t *T, tx ...Transaction) *list[T] {
	return &list[T]{sSelect: SelectContext(ctx, t, tx...).From(t), table: t}
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
	args, values, err := getNonZeroFields(addrMap, l.table, v)
	if err != nil {
		l.err = err
		return l
	}

	helperNonZeroOperation(l.sSelect, args, values)
	return l
}

func (l *list[T]) AsSlice() ([]T, error) {
	if l.err != nil {
		return nil, l.err
	}
	return l.sSelect.AsSlice()
}

func (l *list[T]) AsPagination(page, size uint) (*Pagination[T], error) {
	if l.err != nil {
		return nil, l.err
	}

	return l.sSelect.AsPagination(page, size)
}

func getNonZeroFields[T any](addrMap map[uintptr]field, table *T, value T) ([]any, []any, error) {
	args, values := make([]any, 0), make([]any, 0)

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
				args = append(args, tableOf.Field(i).Addr().Interface())
				values = append(values, valueOf.Field(i).Interface())
			}
		}
	}

	if len(args) == 0 {
		return nil, nil, ErrInvalidArg
	}
	return args, values, nil
}

func helperNonZeroOperation[T any](stateSelect *stateSelect[T], args []any, values []any) {
	stateSelect.Where(equalsOrLike(args[0], values[0]))
	for i := 1; i < len(args); i++ {
		stateSelect.Where(where.Or())
		stateSelect.Where(equalsOrLike(args[i], values[i]))
	}
}

func equalsOrLike(f any, a any) query.Operation {
	v, ok := a.(string)

	if !ok {
		return where.Equals(&f, a)
	}

	if strings.Contains(v, "%") {
		return where.Like(function.ToUpper(f.(*string)), strings.ToUpper(v))
	}

	return where.Equals(&f, a)
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
		if valueOf.Field(i).Kind() != reflect.Pointer {
			//TODO: update to get value from one column query
			return nil, ErrInvalidArg
		}
		fieldOf = valueOf.Field(i).Elem()
		addr := uintptr(fieldOf.Addr().UnsafePointer())
		if addrMap[addr] != nil {
			fields = append(fields, addrMap[addr])
			continue
		}

		if fieldOf.Kind() == reflect.Struct {
			// check if is aggregate
			addr = uintptr(fieldOf.Field(0).Elem().UnsafePointer())
			if addrMap[addr] != nil {
				fields = append(fields, createAggregate(addrMap[addr], fieldOf.Interface()))
				continue
			}
			// check if is function
			addr := uintptr(fieldOf.Field(0).UnsafePointer())
			if addrMap[addr] != nil {
				fields = append(fields, createFunction(addrMap[addr], fieldOf.Interface()))
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

func createFunction(field field, a any) fieldSelect {
	switch a.(type) {
	case query.Function[string]:
		return &functionResult{
			table:         field.table(),
			db:            field.getDb(),
			attributeName: field.getAttributeName(),
			functionType:  enum.UpperFunction}
	}

	return nil
}

func createAggregate(field field, a any) fieldSelect {
	switch a.(type) {
	case query.Count:
		return &aggregateResult{
			table:         field.table(),
			db:            field.getDb(),
			attributeName: field.getAttributeName(),
			aggregateType: enum.CountAggregate}
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

func getArgsTables(builder *builder, addrMap map[uintptr]field, tables []int, args ...any) error {
	if reflect.ValueOf(args[0]).Kind() != reflect.Pointer {
		//TODO: add ErrInvalidTable
		return ErrInvalidArg
	}

	var ptr uintptr
	var i int

	builder.query.Tables = make([]string, 0, len(args))

	valueOf := reflect.ValueOf(args[0]).Elem()
	ptr = uintptr(valueOf.Addr().UnsafePointer())
	if addrMap[ptr] == nil {
		//TODO: add ErrInvalidTable
		return ErrInvalidArg
	}
	tables[i] = addrMap[ptr].getTableId()
	i++
	builder.query.Tables = append(builder.query.Tables, addrMap[ptr].table())

	for _, a := range args[1:] {
		if reflect.ValueOf(a).Kind() != reflect.Pointer {
			//TODO: add ErrInvalidTable
			return ErrInvalidArg
		}

		valueOf = reflect.ValueOf(a).Elem()
		ptr = uintptr(valueOf.Addr().UnsafePointer())
		if addrMap[ptr] == nil {
			//TODO: add ErrInvalidTable
			return ErrInvalidArg
		}
		tables[i] = addrMap[ptr].getTableId()
		i++
		builder.query.Tables = append(builder.query.Tables, addrMap[ptr].table())
	}

	return nil
}

func getArgFunction(arg any, addrMap map[uintptr]field, operation *query.Operation) field {
	value := reflect.ValueOf(arg)
	if value.IsNil() {
		return nil
	}

	if function, ok := value.Elem().Interface().(query.Function[string]); ok {
		operation.Function = function.Type
		return getArg(function.Field, addrMap, nil)
	}
	return getArg(arg, addrMap, nil)
}

func getArg(arg any, addrMap map[uintptr]field, operation *query.Operation) field {
	v := reflect.ValueOf(arg)
	if v.Kind() != reflect.Pointer {
		return nil
	}

	if operation != nil {
		return getArgFunction(arg, addrMap, operation)
	}

	addr := uintptr(v.UnsafePointer())
	if addrMap[addr] != nil {
		return addrMap[addr]
	}
	// any as pointer, used on save, find, remove and list
	return getAnyArg(v, addrMap)
}

// used only inside getArg
func getAnyArg(value reflect.Value, addrMap map[uintptr]field) field {
	if value.IsNil() {
		return nil
	}

	value = reflect.ValueOf(value.Elem().Interface())
	if value.Kind() != reflect.Pointer {
		return nil
	}

	addr := uintptr(value.UnsafePointer())
	if addrMap[addr] != nil {
		return addrMap[addr]
	}
	return nil
}

func helperWhere(builder *builder, addrMap map[uintptr]field, brs ...query.Operation) error {
	for _, br := range brs {
		switch br.Type {
		case enum.OperationWhere:
			if a := getArg(br.Arg, addrMap, &br); a != nil {
				br.Table = a.table()
				br.Attribute = a.getAttributeName()

				builder.brs = append(builder.brs, br)
				continue
			}
			return ErrInvalidWhere
		case enum.OperationAttributeWhere:
			if a, b := getArg(br.Arg, addrMap, nil), getArg(br.Value.GetValue(), addrMap, nil); a != nil && b != nil {
				br.Table = a.table()
				br.Attribute = a.getAttributeName()

				br.AttributeValue = b.getAttributeName()
				br.AttributeValueTable = b.table()
				builder.brs = append(builder.brs, br)
				continue
			}
			return ErrInvalidWhere
		case enum.OperationInWhere:
			if a := getArg(br.Arg, addrMap, &br); a != nil {
				br.Table = a.table()
				br.Attribute = a.getAttributeName()

				builder.brs = append(builder.brs, br)
				continue
			}
			return ErrInvalidWhere
		case enum.OperationIsWhere:
			if a := getArg(br.Arg, addrMap, nil); a != nil {
				br.Table = a.table()
				br.Attribute = a.getAttributeName()

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
