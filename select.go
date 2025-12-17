package goe

import (
	"context"
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

type stateSelect[T any] struct {
	conn    model.Connection
	builder builder
	ctx     context.Context
	argsSelect
}

type find[T any] struct {
	sSelect stateSelect[T]
}

// Find returns a matched record,
// if non record is found returns a [ErrNotFound].
//
// Find uses [context.Background] internally;
// to specify the context, use [FindContext].
//
// # Example
//
//	// one primary key
//	animal, err = goe.Find(db.Animal).ByValue(Animal{ID: 2})
//
//	// two primary keys
//	animalFood, err = goe.Find(db.AnimalFood).ByValue(AnimalFood{AnimalID: 3, FoodID: 2})
//
//	// find record by value, if have more than one it will returns the first
//	cat, err = goe.Find(db.Animal).ByValue(Animal{Name: "Cat"})
func Find[T any](table *T) find[T] {
	return FindContext(context.Background(), table)
}

// FindContext returns a matched record,
// if non record is found returns a [ErrNotFound].
//
// See [Find] for examples
func FindContext[T any](ctx context.Context, table *T) find[T] {
	return find[T]{sSelect: ListContext(ctx, table)}
}

// OnTransaction sets a transaction on the query.
//
// # Example
//
//	tx, err = db.NewTransaction()
//	if err != nil {
//		// handler error
//	}
//	defer tx.Rollback()
//
//	var animals []Animal
//
//	animals, err = goe.List(db.Animal).OnTransaction(tx).AsSlice()
//	if err != nil {
//		// handler error
//	}
//
//	err = tx.Commit()
//	if err != nil {
//		// handler error
//	}
func (f find[T]) OnTransaction(tx model.Transaction) find[T] {
	f.sSelect = f.sSelect.OnTransaction(tx)
	return f
}

// Finds the record by non-zero values,
// if returns more than one it's returns the first
// and ignores the rest
func (f find[T]) ByValue(value T) (*T, error) {
	pks, valuesPks, skip := getNonZeroFields(getArgs{
		addrMap:   addrMap.mapField,
		tableArgs: f.sSelect.tableArgs,
		value:     value})

	if skip {
		return nil, ErrNotFound
	}

	f.sSelect = f.sSelect.Where(operations(pks, valuesPks))

	for row, err := range f.sSelect.Rows() {
		if err != nil {
			return nil, err
		}
		return &row, nil
	}

	return nil, ErrNotFound
}

// Select retrieves rows from tables.
//
// Select uses [context.Background] internally;
// to specify the context, use [SelectContext]
//
// # Example
//
//	var result []struct {
//		User    string
//		Role    *string
//		EndTime *time.Time
//	}
//
//	// row is the generic struct
//	for row, err := range goe.Select[struct {
//			User    string     // output row
//			Role    *string    // output row
//			EndTime *time.Time // output row
//		}](&db.User.Name, &db.Role.Name, &db.UserRole.EndDate).
//		Joins(
//			join.LeftJoin[int](&db.User.ID, &db.UserRole.UserID),
//			join.LeftJoin[int](&db.UserRole.RoleID, &db.Role.ID),
//		).
//		OrderByAsc(&db.User.ID).Rows() {
//
//		if err != nil {
//			//handler error
//		}
//		//handler rows
//		result = append(result, row)
//	}
func Select[T any](args ...any) stateSelect[T] {
	return SelectContext[T](context.Background(), args...)
}

// SelectContext retrieves rows from tables.
//
// See [Select] for examples
func SelectContext[T any](ctx context.Context, args ...any) stateSelect[T] {
	var state stateSelect[T] = createSelectState[T](ctx, getArgsSelect, args...)
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

// Match creates a where on non-zero values over the model T, all strings will be a LIKE operator
// using the ToUpper function to ensure all values is matched.
func (s stateSelect[T]) Match(value T) stateSelect[T] {
	args, values, skip := getNonZeroFields(getArgs{
		addrMap:   addrMap.mapField,
		tableArgs: s.tableArgs,
		value:     value})

	if skip {
		return s
	}

	return s.Filter(operationsList(args, values))
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

// OrderByAsc makes a ordained by args ascending query
func (s stateSelect[T]) OrderByAsc(args ...any) stateSelect[T] {
	for _, arg := range args {
		if a, ok := getAttribute(arg, addrMap.mapField); ok {
			s.builder.query.OrderBy = append(s.builder.query.OrderBy, model.OrderBy{Attribute: a})
		}
	}
	return s
}

// OrderByDesc makes a ordained by args descending query
func (s stateSelect[T]) OrderByDesc(args ...any) stateSelect[T] {
	for _, arg := range args {
		if a, ok := getAttribute(arg, addrMap.mapField); ok {
			s.builder.query.OrderBy = append(s.builder.query.OrderBy, model.OrderBy{Attribute: a, Desc: true})
		}
	}
	return s
}

// GroupBy makes a group by args
func (s stateSelect[T]) GroupBy(args ...any) stateSelect[T] {
	s.builder.query.GroupBy = make([]model.GroupBy, len(args))
	for i := range args {
		if a, ok := getAttribute(args[i], addrMap.mapField); ok {
			s.builder.query.GroupBy[i].Attribute = a
		}
	}
	return s
}

// Joins receives [model.Joins] as joins from join sub package
//
// Deprecated: Use Join, LeftJoin or RightJoin instead.
func (s stateSelect[T]) Joins(joins ...model.Joins) stateSelect[T] {
	for _, j := range joins {
		s.builder.buildSelectJoins(j.Join(), getArgsJoin(addrMap.mapField, j.FirstArg(), j.SecondArg()))
	}
	return s
}

func (s stateSelect[T]) Join(left, right any) stateSelect[T] {
	s.builder.buildSelectJoins(enum.Join, getArgsJoin(addrMap.mapField, left, right))
	return s
}

func (s stateSelect[T]) LeftJoin(left, right any) stateSelect[T] {
	s.builder.buildSelectJoins(enum.LeftJoin, getArgsJoin(addrMap.mapField, left, right))
	return s
}

func (s stateSelect[T]) RightJoin(left, right any) stateSelect[T] {
	s.builder.buildSelectJoins(enum.RightJoin, getArgsJoin(addrMap.mapField, left, right))
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
	stateCount := Select[struct{ Count int64 }](aggregate.Count(s.tableArgs[0]))

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
		count = row.Count
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

// OnTransaction sets a transaction on the query.
//
// # Example
//
//	tx, err = db.NewTransaction()
//	if err != nil {
//		// handler error
//	}
//	defer tx.Rollback()
//
//	var animals []Animal
//
//	animals, err = goe.List(db.Animal).OnTransaction(tx).AsSlice()
//	if err != nil {
//		// handler error
//	}
//
//	err = tx.Commit()
//	if err != nil {
//		// handler error
//	}
func (s stateSelect[T]) OnTransaction(tx model.Transaction) stateSelect[T] {
	s.builder.query.ForUpdate = true
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

func createSelectState[T any](ctx context.Context, getArgs func(args ...any) argsSelect, args ...any) stateSelect[T] {
	s := stateSelect[T]{builder: createBuilder(enum.SelectQuery), argsSelect: getArgs(args...), ctx: ctx}
	s.builder.fieldsSelect = s.fields
	return s
}

// List is a wrapper over [Select] for more simple queries using filters, pagination and ordering.
//
// List uses [context.Background] internally;
// to specify the context, use [ListContext]
//
// # Example
//
//	// where animals.name LIKE $1 AND animal.id = $2 AND animals.habitat_id = $3
//	goe.List(db.Animal).OrderByAsc(&db.Animal.Name).Match(Animal{Name: "Cat", Id: 3, HabitatID: &habitatId}).AsSlice()
//
//	// pagination list
//	var p *goe.Pagination[Animal]
//	p, err = goe.List(db.Animal).AsPagination(1, 10)
func List[T any](table *T) stateSelect[T] {
	return ListContext(context.Background(), table)
}

// ListContext is a wrapper over [Select] for more simple queries using filters, pagination and ordering.
//
// See [List] for examples.
func ListContext[T any](ctx context.Context, table *T) stateSelect[T] {
	return createSelectState[T](ctx, getArgsList, table)
}

type getArgs struct {
	addrMap   map[uintptr]field
	value     any
	tableArgs []any
}

func getNonZeroFields(a getArgs) ([]any, []any, bool) {
	args, values := make([]any, 0), make([]any, 0)

	valueOf := reflect.ValueOf(a.value)
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			args = append(args, a.tableArgs[i])
			values = append(values, valueOf.Field(i).Interface())
		}
	}

	if len(args) == 0 {
		return nil, nil, true
	}
	return args, values, false
}

func operations(args, values []any) model.Operation {
	if len(args) == 1 {
		return equals(args[0], values[0])
	}

	if len(args) == 2 {
		return where.And(equals(args[0], values[0]), equals(args[1], values[1]))
	}

	middle := len(args) / 2

	return where.And(operations(args[:middle], values[:middle]), operations(args[middle:], values[middle:]))
}

func equals(f any, a any) model.Operation {
	return where.Equals(&f, a)
}

func operationsList(args, values []any) model.Operation {
	if len(args) == 1 {
		return equalsOrLike(args[0], values[0])
	}

	if len(args) == 2 {
		return where.And(equalsOrLike(args[0], values[0]), equalsOrLike(args[1], values[1]))
	}

	middle := len(args) / 2

	return where.And(operationsList(args[:middle], values[:middle]), operationsList(args[middle:], values[middle:]))
}

func equalsOrLike(f any, a any) model.Operation {
	v, ok := a.(string)

	if !ok {
		return where.Equals(&f, a)
	}

	return where.Like(function.ToUpper(f.(*string)), strings.ToUpper("%"+v+"%"))
}

type argsSelect struct {
	fields    []fieldSelect
	tableArgs []any
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
	var valueOf reflect.Value
	var f field
	for i := range args {
		valueOf = reflect.ValueOf(args[i])
		if valueOf.Kind() == reflect.Pointer {
			ptr = uintptr(valueOf.UnsafePointer())
			f = addrMap[ptr]
			if f != nil {
				fields[i] = f
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

func getAttribute(arg any, addrMap map[uintptr]field) (model.Attribute, bool) {
	v := reflect.ValueOf(arg)
	if v.Kind() != reflect.Pointer {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	f := addrMap[uintptr(v.UnsafePointer())]
	if f != nil {
		return model.Attribute{Table: f.table(), Name: f.getAttributeName()}, true
	}

	if a, ok := v.Elem().Interface().(model.Attributer); ok {
		f = addrMap[uintptr(reflect.ValueOf(a.GetField()).UnsafePointer())]
		if f != nil {
			return a.Attribute(model.Body{
				Table: f.table(),
				Name:  f.getAttributeName(),
			}), true
		}
	}

	return model.Attribute{}, false
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

func getArgsSelect(args ...any) argsSelect {
	addrMap := addrMap.mapField
	fields := make([]fieldSelect, 0, len(args))

	for _, arg := range args {
		fieldOf := reflect.ValueOf(arg)
		f := addrMap[uintptr(fieldOf.UnsafePointer())]
		if f != nil {
			fields = append(fields, f)
			continue
		}
		if a, ok := fieldOf.Interface().(model.Attributer); ok {
			f = addrMap[uintptr(reflect.ValueOf(a.GetField()).UnsafePointer())]
			if f != nil {
				if a.Attribute(model.Body{}).AggregateType != 0 {
					fields = append(fields, createAggregate(f, fieldOf.Elem().Interface()))
					continue
				}
				fields = append(fields, createFunction(f, fieldOf.Elem().Interface()))
			}
		}
	}

	if len(fields) == 0 {
		panic("goe: invalid argument. try sending a pointer to a database mapped argument")
	}

	return argsSelect{fields: fields, tableArgs: args}
}

func getArgsList(args ...any) argsSelect {
	addrMap := addrMap.mapField
	fields := make([]fieldSelect, 0, len(args))
	tableArgs := make([]any, 0, len(args))

	for _, arg := range args {
		structOf := reflect.ValueOf(arg).Elem()
		var fieldOf reflect.Value
		for i := 0; i < structOf.NumField(); i++ {
			fieldOf = structOf.Field(i)
			if f := addrMap[uintptr(fieldOf.Addr().UnsafePointer())]; f != nil {
				fields = append(fields, f)
				tableArgs = append(tableArgs, fieldOf.Addr().Interface())
			}
		}
	}

	if len(fields) == 0 {
		panic("goe: invalid argument. try sending a pointer to a database mapped argument")
	}

	return argsSelect{fields: fields, tableArgs: tableArgs}
}
