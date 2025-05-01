package goe

import (
	"context"
	"reflect"

	"github.com/go-goe/goe/enum"
	"github.com/go-goe/goe/model"
)

type save[T any] struct {
	table         *T
	errBadRequest error
	update        stateUpdate[T]
}

// Save is a wrapper over [Update] for more simple updates,
// uses the value for create a where matching the primary keys
// and includes for update all non-zero values excluding the primary keys.
// If the record don't exists returns a [ErrNotFound].
//
// Save uses [context.Background] internally;
// to specify the context, use [SaveContext].
//
// # Examples
//
//	// updates animal name on record id 1
//	err = goe.Save(db.Animal).ByValue(Animal{Id: 1, Name: "Cat"})
func Save[T any](table *T) save[T] {
	return SaveContext(context.Background(), table)
}

// SaveContext is a wrapper over [Update] for more simple updates,
// uses the value for create a where matching the primary keys
// and includes for update all non-zero values excluding the primary keys.
//
// See [Save] for examples.
func SaveContext[T any](ctx context.Context, table *T) save[T] {
	return save[T]{update: UpdateContext(ctx, table), table: table, errBadRequest: ErrBadRequest}
}

func (s save[T]) OnTransaction(tx Transaction) save[T] {
	s.update.conn = tx
	return s
}

// Replace the ErrBadRequest with err
func (s save[T]) OnErrBadRequest(err error) save[T] {
	s.errBadRequest = err
	return s
}

func (s save[T]) ByValue(v T) error {
	argsSave := getArgsSave(addrMap.mapField, s.table, v, s.errBadRequest)
	if argsSave.err != nil {
		return argsSave.err
	}

	s.update.builder.sets = argsSave.sets
	return s.update.Where(operations(argsSave.argsWhere, argsSave.valuesWhere))
}

type stateUpdate[T any] struct {
	conn    Connection
	builder builder
	ctx     context.Context
}

// Update updates records in the given table
//
// Update uses [context.Background] internally;
// to specify the context, use [UpdateContext].
//
// # Examples
//
//	// update only the attribute IdJobTitle from PersonJobTitle with the value 3
//	err = goe.Update(db.PersonJobTitle).
//	Sets(update.Set(&db.PersonJobTitle.IdJobTitle, 3)).
//	Where(
//		where.And(
//			where.Equals(&db.PersonJobTitle.PersonId, 2),
//			where.Equals(&db.PersonJobTitle.IdJobTitle, 1),
//	    ),
//	)
//
//	// update all animals name to Cat
//	goe.Update(db.Animal).Sets(update.Set(&db.Animal.Name, "Cat")).All()
func Update[T any](table *T) stateUpdate[T] {
	return UpdateContext(context.Background(), table)
}

// Update updates records in the given table
//
// See [Update] for examples
func UpdateContext[T any](ctx context.Context, table *T) stateUpdate[T] {
	return createUpdateState[T](ctx)
}

// Sets one or more arguments for update
func (s stateUpdate[T]) Sets(sets ...model.Set) stateUpdate[T] {
	for i := range sets {
		s.builder.sets = append(s.builder.sets, set{attribute: getArg(sets[i].Attribute, addrMap.mapField, nil), value: sets[i].Value})
	}

	return s
}

func (s stateUpdate[T]) OnTransaction(tx Transaction) stateUpdate[T] {
	s.conn = tx
	return s
}

// Update all records
func (s stateUpdate[T]) All() error {
	return s.Where(model.Operation{})
}

// Where receives [model.Operation] as where operations from where sub package
func (s stateUpdate[T]) Where(o model.Operation) error {
	helperWhere(&s.builder, addrMap.mapField, o)

	s.builder.buildUpdate()

	driver := s.builder.sets[0].attribute.getDb().driver
	if s.conn == nil {
		s.conn = driver.NewConnection()
	}

	return handlerValues(s.ctx, s.conn, s.builder.query, driver.GetDatabaseConfig())
}

type argSave struct {
	sets        []set
	argsWhere   []any
	valuesWhere []any
	err         error
}

func getArgsSave[T any](addrMap map[uintptr]field, table *T, value T, errBadRequest error) argSave {
	if table == nil {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	tableOf := reflect.ValueOf(table).Elem()

	if tableOf.Kind() != reflect.Struct {
		panic("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
	}

	valueOf := reflect.ValueOf(value)

	sets := make([]set, 0)
	pksWhere, valuesWhere := make([]any, 0, valueOf.NumField()), make([]any, 0, valueOf.NumField())

	var addr uintptr
	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsZero() {
			addr = uintptr(tableOf.Field(i).Addr().UnsafePointer())
			if addrMap[addr] != nil {
				if addrMap[addr].isPrimaryKey() {
					pksWhere = append(pksWhere, tableOf.Field(i).Addr().Interface())
					valuesWhere = append(valuesWhere, valueOf.Field(i).Interface())
					continue
				}
				sets = append(sets, set{attribute: addrMap[addr], value: valueOf.Field(i).Interface()})
			}
		}
	}
	if len(pksWhere) == 0 || len(valuesWhere) == 0 {
		return argSave{err: errBadRequest}
	}
	return argSave{sets: sets, argsWhere: pksWhere, valuesWhere: valuesWhere}
}

func createUpdateState[T any](ctx context.Context) stateUpdate[T] {
	return stateUpdate[T]{builder: createBuilder(enum.UpdateQuery), ctx: ctx}
}
