package goe

import (
	"context"
	"iter"
	"reflect"
	"slices"

	"github.com/go-goe/goe/model"
	"github.com/go-goe/goe/query/update"
	"github.com/go-goe/goe/query/where"
)

type EntityDB[T any] struct {
	isSchema map[int]bool
}

func (e EntityDB[T]) Select(args ...any) iter.Seq2[T, error] {
	argsSelect := getArgsSelectV2(addrMap.mapField, args)

	var s stateSelect[T] = createSelectState[T](context.Background())
	s.builder.fieldsSelect = argsSelect.fields

	s.builder.buildSqlSelect()

	var entity T
	// if query.Header.Err != nil {
	// 	return func(yield func(T, error) bool) {
	// 		yield(entity, dbConfig.ErrorQueryHandler(ctx, query))
	// 	}
	// }
	// dbConfig.InfoHandler(ctx, query)

	dest := make([]any, len(argsSelect.fields))
	value := reflect.ValueOf(&entity).Elem()

	for i := range dest {
		f := (argsSelect.fields[i]).(field)
		if e.isSchema[f.getSchemaID()] {
			value = value.Field(f.getSchemaID())
			if value.IsNil() {
				value.Set(reflect.New(value.Type().Elem()))
			}
			value = value.Elem().Field(f.getEntityID())
			if value.IsNil() {
				value.Set(reflect.New(value.Type().Elem()))
			}
			dest[i] = value.Elem().Field(f.getFieldId()).Addr().Interface()
			continue
		}
		value = value.Field(f.getEntityID())
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		dest[i] = value.Elem().Field(f.getFieldId()).Addr().Interface()
	}

	driver := s.builder.fieldsSelect[0].getDb().driver
	if s.conn == nil {
		s.conn = driver.NewConnection()
	}

	return handlerResultv2(s.ctx, s.conn, s.builder.query, len(s.builder.fieldsSelect), driver.GetDatabaseConfig(), dest, &entity)
}

func getArgsSelectV2(addrMap map[uintptr]field, args []any) argsSelect {

	fields := make([]fieldSelect, 0, len(args))
	var f field
	for _, arg := range args {
		f = addrMap[uintptr(reflect.ValueOf(arg).UnsafePointer())]
		if f != nil {
			fields = append(fields, f)
		}
	}

	if len(fields) == 0 {
		panic("goe: invalid arguments. try sending a pointer to a database argument")
	}

	return argsSelect{fields: fields, tableArgs: args}
}

type Entity[T any] struct {
	entity     *T
	entityArgs []any
	operation  model.Operation
}

func NewEntity[T any](e *T) Entity[T] {
	entityOf := reflect.ValueOf(e).Elem()
	args := make([]any, entityOf.NumField()-1)
	for i := range entityOf.NumField() - 1 {
		args[i] = entityOf.Field(i).Addr().Interface()
	}

	return Entity[T]{entity: e, entityArgs: args}
}

func (e Entity[T]) Equals(value T) wherer[T] {
	valueOf := reflect.ValueOf(value)
	var fieldOf reflect.Value
	values := make([]any, 0, len(e.entityArgs))
	args := make([]any, 0, len(e.entityArgs))
	for i := range len(e.entityArgs) {
		fieldOf = valueOf.Field(i)
		if !fieldOf.IsZero() {
			values = append(values, fieldOf.Interface())
			args = append(args, e.entityArgs[i])
		}
	}

	if e.operation.Operator == 0 {
		e.operation = where.Equals(&args[0], values[0])
		for i := 1; i < len(values); i++ {
			e.operation = where.And(e.operation, where.Equals(&args[i], values[i]))
		}
		return e
	}

	for i := 0; i < len(values); i++ {
		e.operation = where.And(e.operation, where.Equals(&args[i], values[i]))
	}

	return e
}

func (e Entity[T]) Find(value T) (*T, error) {
	return e.FindContext(context.Background(), value)
}

func (e Entity[T]) FindContext(ctx context.Context, value T) (*T, error) {
	return FindContext(ctx, e.entity).ByValue(value)
}

func (e Entity[T]) Save(value T) error {
	return e.SaveContext(context.Background(), value)
}

func (e Entity[T]) SaveContext(ctx context.Context, value T) error {
	return SaveContext(ctx, e.entity).ByID(value)
}

func (e Entity[T]) Create(value *T) error {
	return e.CreateContext(context.Background(), value)
}

func (e Entity[T]) CreateContext(ctx context.Context, value *T) error {
	return InsertContext(ctx, e.entity).One(value)
}

func (e Entity[T]) CreateAll(values []T) error {
	return e.CreateAllContext(context.Background(), values)
}

func (e Entity[T]) CreateAllContext(ctx context.Context, values []T) error {
	return InsertContext(ctx, e.entity).All(values)
}

func (e Entity[T]) Remove(value T) error {
	return RemoveContext(context.Background(), e.entity).ByValue(value)
}

func (e Entity[T]) RemoveContext(ctx context.Context, value T) error {
	return RemoveContext(ctx, e.entity).ByValue(value)
}

// ============= finalizer =============

func (e Entity[T]) List() ([]T, error) {
	return e.ListContext(context.Background())
}

func (e Entity[T]) ListContext(ctx context.Context) ([]T, error) {
	return SelectContext[T](ctx, e.entity).Where(e.operation).AsSlice()
}

func (e Entity[T]) Update(value T, uses ...any) error {
	return e.UpdateContext(context.Background(), value, uses...)
}

func (e Entity[T]) UpdateContext(ctx context.Context, value T, uses ...any) error {
	valueOf := reflect.ValueOf(value)
	var fieldOf reflect.Value
	values := make([]any, 0, len(e.entityArgs))
	args := make([]any, 0, len(e.entityArgs))
	for i := range len(e.entityArgs) {
		fieldOf = valueOf.Field(i)
		if !fieldOf.IsZero() || slices.ContainsFunc(uses, func(use any) bool { return use == e.entityArgs[i] }) {
			values = append(values, fieldOf.Interface())
			args = append(args, e.entityArgs[i])
		}
	}

	if len(values) == 0 {
		return nil
	}

	sets := make([]model.Set, len(values))
	for i := 0; i < len(values); i++ {
		sets[i] = update.Set(&args[i], values[i])
	}

	return UpdateContext(ctx, e.entity).Sets(sets...).Where(e.operation)
}

func (e Entity[T]) Delete() error {
	return e.DeleteContext(context.Background())
}

func (e Entity[T]) DeleteContext(ctx context.Context) error {
	return DeleteContext(ctx, e.entity).Where(e.operation)
}
