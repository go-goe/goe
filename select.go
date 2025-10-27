package goe

import (
	"context"
	"errors"
	"iter"
	"math"
	"reflect"
	"slices"
	"strings"

	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
	"github.com/go-goe/goe/query"
	"github.com/go-goe/goe/query/aggregate"
	"github.com/go-goe/goe/query/function"
	"github.com/go-goe/goe/query/where"
)

var ErrNotFound = errors.New("goe: not found any element on result set")

type stateSelect[T any] struct {
	conn    Connection
	builder builder
	ctx     context.Context
	table   any
}

type find[T any] struct {
	table       *T
	errNotFound error
	sSelect     stateSelect[T]
}

// Find returns a matched record,
// if non record is found returns a [ErrNotFound].
//
// Find uses [context.Background] internally;
// to specify the context, use [FindContext].
//
// # Example
//
//	goe.Find(db.Animal).ByID(Animal{Id: 2})
func Find[T any](table *T) find[T] {
	return FindContext(context.Background(), table)
}

// FindContext returns a matched record,
// if non record is found returns a [ErrNotFound].
//
// See [Find] for examples
func FindContext[T any](ctx context.Context, table *T) find[T] {
	return find[T]{table: table, sSelect: SelectContext[T](ctx, table), errNotFound: ErrNotFound}
}

func (f find[T]) OnTransaction(tx Transaction) find[T] {
	f.sSelect.conn = tx
	return f
}

// Replace the ErrNotFound with err
func (f find[T]) OnErrNotFound(err error) find[T] {
	f.errNotFound = err
	return f
}

// Finds the record by values on Ids
func (f find[T]) ByID(value T) (*T, error) {
	pks, valuesPks, err := getArgsPks(getArgs{
		addrMap: addrMap.mapField,
		table:   f.table,
		value:   value})

	if err != nil {
		return nil, err
	}

	f.sSelect = f.sSelect.Where(operations(pks, valuesPks))

	for row, err := range f.sSelect.Rows() {
		if err != nil {
			return nil, err
		}
		return &row, nil
	}

	return nil, f.errNotFound
}

// Finds the record by non-zero values,
// if returns more than one it's returns the first
// and ignores the rest
func (f find[T]) ByValue(value T) (*T, error) {
	pks, valuesPks, skip := getNonZeroFields(getArgs{
		addrMap: addrMap.mapField,
		table:   f.table,
		value:   value})

	if skip {
		return nil, f.errNotFound
	}

	f.sSelect = f.sSelect.Where(operations(pks, valuesPks))

	for row, err := range f.sSelect.Rows() {
		if err != nil {
			return nil, err
		}
		return &row, nil
	}

	return nil, f.errNotFound
}

// Select retrieves rows from tables.
//
// Select uses [context.Background] internally;
// to specify the context, use [SelectContext]
//
// # Example
//
//	// simple select
//	goe.Select(db.Animal).AsSlice()
//
//	// iterator select
//	for row, err := range goe.Select(db.Animal).Rows() { ... }
//
//	// pagination select
//	var p *goe.Pagination[Animal]
//	p, err = goe.Select(db.Animal).AsPagination(1, 10)
//
//	// select with where, joins and order by
//	goe.Select(db.Food).
//		Joins(
//			join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
//			join.Join[int](&db.AnimalFood.IdAnimal, &db.Animal.Id),
//			join.Join[uuid.UUID](&db.Animal.IdHabitat, &db.Habitat.Id),
//			join.Join[int](&db.Habitat.IdWeather, &db.Weather.Id),
//		).
//		Where(
//			where.And(
//				where.Equals(&db.Food.Id, 1),
//				where.Equals(&db.Food.Name, "Beef"),
//			),
//		).OrderByAsc(&db.Food.Name).AsSlice()
//
//	// select any argument
//	goe.Select(&struct {
//		User    *string
//		Role    *string
//		EndTime **time.Time
//	}{
//		User:    &db.User.Name,
//		Role:    &db.Role.Name,
//		EndTime: &db.UserRole.EndDate,
//	}).AsSlice()
func Select[T any](table any) stateSelect[T] {
	return SelectContext[T](context.Background(), table)
}

// SelectContext retrieves rows from tables.
//
// See [Select] for examples
func SelectContext[T any](ctx context.Context, table any) stateSelect[T] {
	var state stateSelect[T] = createSelectState[T](ctx)
	argsSelect := getArgsSelect(addrMap.mapField, table)

	state.table = argsSelect.table
	state.builder.fieldsSelect = argsSelect.fields
	return state
}

// Where receives [model.Operation] as where operations from where sub package
func (s stateSelect[T]) Where(o model.Operation) stateSelect[T] {
	s.builder.brs = nil
	helperWhere(&s.builder, addrMap.mapField, o)
	return s
}

// Filter creates a where on non-zero values.
func (s stateSelect[T]) Filter(o model.Operation) stateSelect[T] {
	s.builder.filters = nil
	helperFilter(&s.builder, addrMap.mapField, o)
	return s
}

// Take takes i elements
func (s stateSelect[T]) Take(i int) stateSelect[T] {
	s.builder.query.Limit = i
	return s
}

// Skip skips i elements
func (s stateSelect[T]) Skip(i int) stateSelect[T] {
	s.builder.query.Offset = i
	return s
}

// OrderByAsc makes a ordained by arg ascending query
func (s stateSelect[T]) OrderByAsc(arg any) stateSelect[T] {
	field := getArg(arg, addrMap.mapField, nil)
	s.builder.query.OrderBy = &model.OrderBy{Attribute: model.Attribute{Name: field.getAttributeName(), Table: field.table()}}
	return s
}

// OrderByDesc makes a ordained by arg descending query
func (s stateSelect[T]) OrderByDesc(arg any) stateSelect[T] {
	field := getArg(arg, addrMap.mapField, nil)
	s.builder.query.OrderBy = &model.OrderBy{
		Attribute: model.Attribute{Name: field.getAttributeName(), Table: field.table()},
		Desc:      true}
	return s
}

// Joins receives [model.Joins] as joins from join sub package
func (s stateSelect[T]) Joins(joins ...model.Joins) stateSelect[T] {
	for _, j := range joins {
		s.builder.buildSelectJoins(j.Join(), getArgsJoin(addrMap.mapField, j.FirstArg(), j.SecondArg()))
	}
	return s
}

// AsSlice return all the rows as a slice.
func (s stateSelect[T]) AsSlice() ([]T, error) {
	rows := make([]T, 0, s.builder.query.Limit)
	for row, err := range s.Rows() {
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// AsQuery return a [model.Query] for use inside a [where.In].
func (s stateSelect[T]) AsQuery() model.Query {
	s.builder.buildSqlSelect()
	return s.builder.query
}

type Pagination[T any] struct {
	TotalValues int64 `json:"totalValues"`
	TotalPages  int   `json:"totalPages"`

	PageValues int `json:"pageValues"`
	PageSize   int `json:"pageSize"`

	CurrentPage     int  `json:"currentPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
	PreviousPage    int  `json:"previousPage"`
	HasNextPage     bool `json:"hasNextPage"`
	NextPage        int  `json:"nextPage"`

	StartIndex int `json:"startIndex"`
	EndIndex   int `json:"endIndex"`
	Values     []T `json:"values"`
}

// AsPagination return a paginated query as [Pagination].
//
// Default values for page and size are 1 and 10 respectively.
func (s stateSelect[T]) AsPagination(page, size int) (*Pagination[T], error) {
	if size <= 0 {
		size = 10
	}
	if page <= 0 {
		page = 1
	}

	var err error
	stateCount := Select[struct{ query.Count }](&struct {
		*query.Count
	}{
		Count: aggregate.Count(s.table),
	})

	// copy joins
	stateCount.builder.joins = s.builder.joins
	stateCount.builder.joinsArgs = s.builder.joinsArgs

	// copy operations
	stateCount.builder.brs = s.builder.brs
	stateCount.builder.filters = s.builder.filters

	// copy connection/transaction
	stateCount.conn = s.conn

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

	p.TotalPages = int(math.Ceil(float64(count) / float64(size)))
	p.CurrentPage = page

	if page == p.TotalPages || p.TotalPages == 0 {
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
		p.EndIndex = int(p.TotalValues)
	} else {
		p.EndIndex = size * page
	}

	return p, nil
}

func (s stateSelect[T]) OnTransaction(tx Transaction) stateSelect[T] {
	s.conn = tx
	return s
}

// Rows return a iterator on rows.
func (s stateSelect[T]) Rows() iter.Seq2[T, error] {
	s.builder.buildSqlSelect()

	driver := s.builder.fieldsSelect[0].getDb().driver
	if s.conn == nil {
		s.conn = driver.NewConnection()
	}

	return handlerResult[T](s.ctx, s.conn, s.builder.query, len(s.builder.fieldsSelect), driver.GetDatabaseConfig())
}

func createSelectState[T any](ctx context.Context) stateSelect[T] {
	return stateSelect[T]{builder: createBuilder(enum.SelectQuery), ctx: ctx}
}

type list[T any] struct {
	table   *T
	sSelect stateSelect[T]
	err     error
}

// List is a wrapper over [Select] for more simple queries using filters, pagination and ordering.
//
// List uses [context.Background] internally;
// to specify the context, use [ListContext]
//
// # Example
//
//	// where animals.name LIKE $1
//	// on LIKE Filter goe uses ToUpper to match all results
//	goe.List(db.Animal).OrderByDesc(&db.Animal.Name).Filter(Animal{Name: "%Cat%"}).AsSlice()
//
//	// where animals.name equals $1 AND animal.id = $2 AND animals.idhabitat = $3
//	goe.List(db.Animal).OrderByAsc(&db.Animal.Name).Filter(Animal{Name: "Cat", Id: animals[0].Id, IdHabitat: &habitats[0].Id}).AsSlice()
//
//	// pagination list
//	var p *goe.Pagination[Animal]
//	p, err = goe.List(db.Animal).AsPagination(1, 10)
func List[T any](table *T) list[T] {
	return ListContext(context.Background(), table)
}

// ListContext is a wrapper over [Select] for more simple queries using filters, pagination and ordering.
//
// See [List] for examples.
func ListContext[T any](ctx context.Context, table *T) list[T] {
	return list[T]{sSelect: SelectContext[T](ctx, table), table: table}
}

// OrderByAsc makes a ordained by arg ascending query.
func (l list[T]) OrderByAsc(a any) list[T] {
	l.sSelect = l.sSelect.OrderByAsc(a)
	return l
}

// OrderByDesc makes a ordained by arg descending query.
func (l list[T]) OrderByDesc(a any) list[T] {
	l.sSelect = l.sSelect.OrderByDesc(a)
	return l
}

// Filter creates a where on non-zero values.
func (l list[T]) Filter(o model.Operation) list[T] {
	l.sSelect = l.sSelect.Filter(o)
	return l
}

func (l list[T]) Where(o model.Operation) list[T] {
	l.sSelect = l.sSelect.Where(o)
	return l
}

func (l list[T]) Joins(joins ...model.Joins) list[T] {
	l.sSelect = l.sSelect.Joins(joins...)
	return l
}

func (l list[T]) OnTransaction(tx Transaction) list[T] {
	l.sSelect.conn = tx
	return l
}

// AsSlice return all the rows as a slice
func (l list[T]) AsSlice() ([]T, error) {
	if l.err != nil {
		return nil, l.err
	}
	return l.sSelect.AsSlice()
}

// AsPagination return a paginated query as [Pagination].
//
// Default values for page and size are 1 and 10 respectively.
func (l list[T]) AsPagination(page, size int) (*Pagination[T], error) {
	if l.err != nil {
		return nil, l.err
	}

	return l.sSelect.AsPagination(page, size)
}

// Rows return a iterator on rows.
func (l list[T]) Rows() iter.Seq2[T, error] {
	return l.sSelect.Rows()
}

type getArgs struct {
	addrMap map[uintptr]field
	table   any
	value   any
}

func getArgsPks(a getArgs) ([]any, []any, error) {
	args, values := getPrimaryArgs(a)

	if len(args) == 0 {
		return nil, nil, ErrNotFound
	}
	return args, values, nil
}

func getPrimaryArgs(a getArgs) ([]any, []any) {
	if a.table == nil {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	tableOf := reflect.ValueOf(a.table).Elem()

	if tableOf.Kind() != reflect.Struct {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	valueOf := reflect.ValueOf(a.value)

	args, values := make([]any, 0, valueOf.NumField()), make([]any, 0, valueOf.NumField())
	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			if a.addrMap[addr] != nil {
				if a.addrMap[addr].isPrimaryKey() {
					args = append(args, tableOf.Field(i).Addr().Interface())
					values = append(values, valueOf.Field(i).Interface())
				}
			}
		}
	}

	return args, values
}

func getNonZeroFields(a getArgs) ([]any, []any, bool) {
	args, values := make([]any, 0), make([]any, 0)

	tableOf := reflect.ValueOf(a.table).Elem()
	if tableOf.Kind() != reflect.Struct {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	valueOf := reflect.ValueOf(a.value)
	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			if a.addrMap[addr] != nil {
				args = append(args, tableOf.Field(i).Addr().Interface())
				values = append(values, valueOf.Field(i).Interface())
			}
		}
	}

	if len(args) == 0 {
		return nil, nil, true
	}
	return args, values, false
}

func operations(args, values []any) model.Operation {
	if len(args) == 1 {
		return equalsOrLike(args[0], values[0])
	}

	if len(args) == 2 {
		return where.And(equalsOrLike(args[0], values[0]), equalsOrLike(args[1], values[1]))
	}

	middle := len(args) / 2

	return where.And(operations(args[:middle], values[:middle]), operations(args[middle:], values[middle:]))
}

func equalsOrLike(f any, a any) model.Operation {
	v, ok := a.(string)

	if !ok {
		return where.Equals(&f, a)
	}

	if strings.Contains(v, "%") {
		return where.Like(function.ToUpper(f.(*string)), strings.ToUpper(v))
	}

	return where.Equals(&f, a)
}

type argsSelect struct {
	fields []fieldSelect
	table  any
}

func getArgsSelect(addrMap map[uintptr]field, arg any) argsSelect {
	fields := make([]fieldSelect, 0)

	if reflect.ValueOf(arg).Kind() != reflect.Pointer {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	valueOf := reflect.ValueOf(arg).Elem()

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
			continue
		}
		//get args from anonymous struct
		return getArgsSelectAno(addrMap, valueOf)
	}

	if len(fields) == 0 {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	return argsSelect{fields: fields, table: arg}
}

func getArgsSelectAno(addrMap map[uintptr]field, valueOf reflect.Value) argsSelect {
	fields := make([]fieldSelect, 0)
	var fieldOf reflect.Value
	var table any = valueOf.Field(0).Elem().Addr().Interface()
	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).Kind() != reflect.Pointer {
			//TODO: update to get value from one column query
			panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
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
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	if len(fields) == 0 {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	return argsSelect{fields: fields, table: table}
}

func createFunction(field field, a any) fieldSelect {
	if f, ok := a.(model.FunctionType); ok {
		return functionResult{
			tableName:     field.table(),
			schemaName:    field.schema(),
			tableId:       field.getTableId(),
			db:            field.getDb(),
			attributeName: field.getAttributeName(),
			functionType:  f.GetType()}
	}

	return nil
}

func createAggregate(field field, a any) fieldSelect {
	if ag, ok := a.(model.Aggregate); ok {
		return aggregateResult{
			tableName:     field.table(),
			schemaName:    field.schema(),
			tableId:       field.getTableId(),
			db:            field.getDb(),
			attributeName: field.getAttributeName(),
			aggregateType: ag.Aggregate()}
	}

	return nil
}

func getArgsJoin(addrMap map[uintptr]field, args ...any) []field {
	fields := make([]field, 2)
	var ptr uintptr
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			ptr = uintptr(valueOf.Addr().UnsafePointer())
			if addrMap[ptr] != nil {
				fields[i] = addrMap[ptr]
			}
			continue
		}
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	if fields[0] == nil || fields[1] == nil {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	return fields
}

func getArgFunction(arg any, addrMap map[uintptr]field, operation *model.Operation) field {
	value := reflect.ValueOf(arg)
	if value.IsNil() {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	if function, ok := value.Elem().Interface().(query.Function[string]); ok {
		operation.Function = function.Type
		return getArg(function.Field, addrMap, nil)
	}
	return getArg(arg, addrMap, nil)
}

func getArg(arg any, addrMap map[uintptr]field, operation *model.Operation) field {
	v := reflect.ValueOf(arg)
	if v.Kind() != reflect.Pointer {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
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

func helperWhere(builder *builder, addrMap map[uintptr]field, br model.Operation) {
	switch br.Type {
	case enum.OperationWhere, enum.OperationInWhere:
		a := getArg(br.Arg, addrMap, &br)
		br.Table = model.Table{Schema: a.schema(), Name: a.table()}
		br.TableId = a.getTableId()
		br.Attribute = a.getAttributeName()

		builder.brs = append(builder.brs, br)
	case enum.OperationAttributeWhere:
		a, b := getArg(br.Arg, addrMap, nil), getArg(br.Value.GetValue(), addrMap, nil)
		br.Table = model.Table{Schema: a.schema(), Name: a.table()}
		br.TableId = a.getTableId()
		br.Attribute = a.getAttributeName()

		br.AttributeValue = b.getAttributeName()
		br.AttributeValueTable = model.Table{Schema: b.schema(), Name: b.table()}
		br.AttributeTableId = b.getTableId()
		builder.brs = append(builder.brs, br)
	case enum.OperationIsWhere:
		a := getArg(br.Arg, addrMap, nil)
		br.Table = model.Table{Schema: a.schema(), Name: a.table()}
		br.TableId = a.getTableId()
		br.Attribute = a.getAttributeName()

		builder.brs = append(builder.brs, br)
	case enum.LogicalWhere:
		helperWhere(builder, addrMap, *br.FirstOperation)
		builder.brs = append(builder.brs, br)
		helperWhere(builder, addrMap, *br.SecondOperation)
	}
}

func helperFilter(builder *builder, addrMap map[uintptr]field, br model.Operation) bool {
	switch br.Type {
	case enum.OperationWhere, enum.OperationInWhere:
		if !reflect.ValueOf(br.Value.GetValue()).IsZero() {
			a := getArg(br.Arg, addrMap, &br)
			br.Table = model.Table{Schema: a.schema(), Name: a.table()}
			br.TableId = a.getTableId()
			br.Attribute = a.getAttributeName()

			builder.filters = append(builder.filters, br)
			return true
		}
	case enum.OperationAttributeWhere:
		a, b := getArg(br.Arg, addrMap, nil), getArg(br.Value.GetValue(), addrMap, nil)
		br.Table = model.Table{Schema: a.schema(), Name: a.table()}
		br.TableId = a.getTableId()
		br.Attribute = a.getAttributeName()

		br.AttributeValue = b.getAttributeName()
		br.AttributeValueTable = model.Table{Schema: b.schema(), Name: b.table()}
		br.AttributeTableId = b.getTableId()
		builder.filters = append(builder.filters, br)
		return true
	case enum.LogicalWhere:
		firstFlag := helperFilter(builder, addrMap, *br.FirstOperation)
		builder.filters = append(builder.filters, br)
		idx := len(builder.filters) - 1
		secondFlag := helperFilter(builder, addrMap, *br.SecondOperation)
		if !firstFlag || !secondFlag {
			builder.filters = slices.Delete(builder.filters, idx, idx+1)
		}
		return true
	}
	return false
}
