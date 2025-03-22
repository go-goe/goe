package goe

import (
	"context"
	"errors"
	"reflect"
	"slices"

	"github.com/olauro/goe/enum"
	"github.com/olauro/goe/model"
	"github.com/olauro/goe/query/where"
)

type save[T any] struct {
	table         *T
	argsReplace   []any
	valuesReplace []any
	includes      []field
	update        *stateUpdate[T]
}

// Save is a wrapper over [Update] for more simple updates,
// uses the value for create a where matching the primary keys
// and includes for update all non-zero values excluding the primary keys.
//
// Save uses [context.Background] internally;
// to specify the context, use [SaveContext].
//
// # Examples
//
//	// updates animal name on record id 1
//	err = goe.Save(db.Animal).Value(Animal{Id: 1, Name: "Cat"})
//
//	// update all non-zero values including HabitatId
//	// HabitatId can have a zero or nil value and won't be ignored
//	err = goe.Save(db.Animal).Includes(&db.Animal.HabitatId).Value(Animal{Id: 1, Name: "Cat", HabitatId: nil})
//
//	// replace the primary key values from the matched record
//	// updates IdJobTitle from 1 to 3
//	err = goe.Save(db.PersonJobTitle).Replace(PersonJobTitle{
//		PersonId:  2,
//		IdJobTitle: 1}).Value(PersonJobTitle{
//		IdJobTitle: 3, UpdatedAt: time.Now()})
func Save[T any](table *T, tx ...Transaction) *save[T] {
	return SaveContext(context.Background(), table, tx...)
}

// SaveContext is a wrapper over [Update] for more simple updates,
// uses the value for create a where matching the primary keys
// and includes for update all non-zero values excluding the primary keys.
//
// See [Save] for examples.
func SaveContext[T any](ctx context.Context, table *T, tx ...Transaction) *save[T] {
	save := &save[T]{}
	save.update = UpdateContext(ctx, table, tx...)

	if save.update.err != nil {
		return save
	}

	save.table = table

	return save
}

func (s *save[T]) Includes(args ...any) *save[T] {
	ptrArgs, err := getArgsUpdate(addrMap.mapField, args...)
	if err != nil {
		s.update.err = err
		return s
	}

	s.includes = append(s.includes, ptrArgs...)
	return s
}

// Replace is for update a primary key value
func (s *save[T]) Replace(replace T) *save[T] {
	if s.update.err != nil {
		return s
	}

	s.argsReplace, s.valuesReplace, s.update.err = getArgsPks(addrMap.mapField, s.table, replace)
	return s
}

func (s *save[T]) Value(v T) error {
	if s.update.err != nil {
		return s.update.err
	}

	argsSave := getArgsSave(addrMap.mapField, s.table, v)
	if argsSave.err != nil {
		return argsSave.err
	}

	// used for replace primary key model
	if s.argsReplace != nil {
		for i := range argsSave.pks {
			// includes the pk for update if the values are different from replace model
			if !slices.Contains(s.valuesReplace, argsSave.valuesWhere[i]) {
				argsSave.includes = append(argsSave.includes, argsSave.pks[i])
			}
		}
		argsSave.argsWhere = s.argsReplace
		argsSave.valuesWhere = s.valuesReplace
	}

	if len(argsSave.includes) == 0 {
		return errors.New("goe: empty includes. call includes function to include all update fields")
	}

	for i := range s.includes {
		if !slices.ContainsFunc(argsSave.includes, func(f field) bool {
			return f.getFieldId() == s.includes[i].getFieldId()
		}) {
			argsSave.includes = append(argsSave.includes, s.includes[i])
		}
	}

	s.update.Wheres(where.Equals(&argsSave.argsWhere[0], argsSave.valuesWhere[0]))
	for i := 1; i < len(argsSave.argsWhere); i++ {
		s.update.Wheres(where.And())
		s.update.Wheres(where.Equals(&argsSave.argsWhere[i], argsSave.valuesWhere[i]))
	}

	s.update.builder.fields = argsSave.includes
	return s.update.Value(v)
}

type stateUpdate[T any] struct {
	conn    Connection
	builder builder
	ctx     context.Context
	err     error
}

// Update updates records in the given table
//
// Update uses [context.Background] internally;
// to specify the context, use [UpdateContext].
//
// # Examples
//
//	// update only the attribute IdJobTitle from PersonJobTitle with the value 3
//	// the wheres call ensures that only the records that match the query will be updated
//	err = goe.Update(db.PersonJobTitle).Includes(&db.PersonJobTitle.IdJobTitle).Wheres(
//		where.Equals(&db.PersonJobTitle.PersonId, 2),
//		where.And(),
//		where.Equals(&db.PersonJobTitle.IdJobTitle, 1)).
//	Value(PersonJobTitle{IdJobTitle: 3})
func Update[T any](table *T, tx ...Transaction) *stateUpdate[T] {
	return UpdateContext(context.Background(), table, tx...)
}

// Update updates records in the given table
//
// See [Update] for examples
func UpdateContext[T any](ctx context.Context, table *T, tx ...Transaction) *stateUpdate[T] {
	f := getArg(table, addrMap.mapField, nil)

	var state *stateUpdate[T]
	if f == nil {
		state = new(stateUpdate[T])
		state.err = errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
		return state
	}

	db := f.getDb()

	if tx != nil {
		state = createUpdateState[T](tx[0], ctx)
	} else {
		state = createUpdateState[T](db.driver.NewConnection(), ctx)
	}

	return state
}

// Includes one or more args for update
func (s *stateUpdate[T]) Includes(args ...any) *stateUpdate[T] {
	if s.err != nil {
		return s
	}

	fields, err := getArgsUpdate(addrMap.mapField, args...)

	if err != nil {
		s.err = err
		return s
	}

	s.builder.fields = append(s.builder.fields, fields...)
	return s
}

// Wheres receives [model.Operation] as where operations from where sub package
func (s *stateUpdate[T]) Wheres(brs ...model.Operation) *stateUpdate[T] {
	if s.err != nil {
		return s
	}
	s.err = helperWhere(&s.builder, addrMap.mapField, brs...)
	return s
}

func (s *stateUpdate[T]) Value(value T) error {
	if s.err != nil {
		return s.err
	}

	if s.conn == nil {
		return errors.New("goe: empty includes. call includes function to include all update fields")
	}

	v := reflect.ValueOf(value)

	s.err = s.builder.buildSqlUpdate(v)
	if s.err != nil {
		return s.err
	}

	return handlerValues(s.conn, s.builder.query, s.ctx)
}

func getArgsUpdate(addrMap map[uintptr]field, args ...any) ([]field, error) {
	fields := make([]field, 0)
	var valueOf reflect.Value
	for i := range args {
		valueOf = reflect.ValueOf(args[i])

		if valueOf.Kind() != reflect.Pointer {
			return nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
		}

		valueOf = valueOf.Elem()

		if !valueOf.CanAddr() {
			return nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
		}

		addr := uintptr(valueOf.Addr().UnsafePointer())
		if addrMap[addr] != nil {
			fields = append(fields, addrMap[addr])
		}
	}
	if len(fields) == 0 {
		return nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	return fields, nil
}

type argSave struct {
	includes    []field
	pks         []field
	argsWhere   []any
	valuesWhere []any
	err         error
}

func getArgsSave[T any](addrMap map[uintptr]field, table *T, value T) argSave {
	if table == nil {
		return argSave{err: errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")}
	}

	tableOf := reflect.ValueOf(table).Elem()

	if tableOf.Kind() != reflect.Struct {
		return argSave{err: errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")}
	}

	valueOf := reflect.ValueOf(value)

	includes, pks := make([]field, 0), make([]field, 0)
	args, values := make([]any, 0, valueOf.NumField()), make([]any, 0, valueOf.NumField())

	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			if addrMap[addr] != nil {
				if addrMap[addr].isPrimaryKey() {
					pks = append(pks, addrMap[addr])
					args = append(args, tableOf.Field(i).Addr().Interface())
					values = append(values, valueOf.Field(i).Interface())
					continue
				}
				includes = append(includes, addrMap[addr])
			}
		}
	}

	return argSave{includes: includes, pks: pks, argsWhere: args, valuesWhere: values}
}

func getArgsPks[T any](addrMap map[uintptr]field, table *T, value T) ([]any, []any, error) {
	if table == nil {
		return nil, nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	tableOf := reflect.ValueOf(table).Elem()

	if tableOf.Kind() != reflect.Struct {
		return nil, nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	valueOf := reflect.ValueOf(value)

	args, values := make([]any, 0, valueOf.NumField()), make([]any, 0, valueOf.NumField())
	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			if addrMap[addr] != nil {
				if addrMap[addr].isPrimaryKey() {
					args = append(args, tableOf.Field(i).Addr().Interface())
					values = append(values, valueOf.Field(i).Interface())
					continue
				}
			}
		}
	}

	if len(args) == 0 && len(values) == 0 {
		return nil, nil, errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}
	return args, values, nil
}

func createUpdateState[T any](
	conn Connection,
	ctx context.Context) *stateUpdate[T] {
	return &stateUpdate[T]{conn: conn, builder: createBuilder(enum.UpdateQuery), ctx: ctx}
}
