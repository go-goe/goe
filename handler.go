package goe

import (
	"context"
	"iter"
	"reflect"

	"github.com/go-goe/goe/model"
)

func handlerValues(conn Connection, query model.Query, ctx context.Context) error {
	err := conn.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func handlerValuesReturning(conn Connection, query model.Query, value reflect.Value, pkFieldId int, ctx context.Context) error {
	row := conn.QueryRowContext(ctx, query)

	err := row.Scan(value.Field(pkFieldId).Addr().Interface())
	if err != nil {
		return err
	}
	return nil
}

func handlerValuesReturningBatch(conn Connection, query model.Query, value reflect.Value, pkFieldId int, ctx context.Context) error {
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		err = rows.Scan(value.Index(i).Field(pkFieldId).Addr().Interface())
		if err != nil {
			return err
		}
		i++
	}
	return nil
}

func handlerResult[T any](ctx context.Context, conn Connection, query model.Query, numFields int, anonymous bool) iter.Seq2[T, error] {
	rows, err := conn.QueryContext(ctx, query)

	var v T
	if err != nil {
		return func(yield func(T, error) bool) {
			yield(v, err)
		}
	}

	value := reflect.TypeOf(v)
	dest := make([]any, numFields)
	if anonymous {
		fieldElem := make(map[int]bool)
		for i := range dest {
			if value.Field(i).Type.Elem().Kind() == reflect.Pointer {
				fieldElem[i] = true
				dest[i] = reflect.New(value.Field(i).Type.Elem()).Interface()
				continue
			}
			dest[i] = reflect.New(value.Field(i).Type).Interface()
		}
		return mapAnonymousStructQuery[T](rows, dest, value, fieldElem)
	}

	for i := range dest {
		dest[i] = reflect.New(value.Field(i).Type).Interface()
	}

	return mapStructQuery[T](rows, dest, value)
}

func mapStructQuery[T any](rows Rows, dest []any, value reflect.Type) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var (
			err  error
			s, f reflect.Value
		)
		defer rows.Close()
		s = reflect.New(value).Elem()

		for rows.Next() {
			err = rows.Scan(dest...)

			if err != nil {
				yield(s.Interface().(T), err)
				return
			}

			for i, a := range dest {
				f = s.Field(i)
				f.Set(reflect.ValueOf(a).Elem())
			}
			if !yield(s.Interface().(T), err) {
				return
			}
		}
	}
}

func mapAnonymousStructQuery[T any](rows Rows, dest []any, value reflect.Type, fieldMap map[int]bool) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var (
			err  error
			s, f reflect.Value
		)
		defer rows.Close()
		s = reflect.New(value).Elem()

		for rows.Next() {
			err = rows.Scan(dest...)

			if err != nil {
				yield(s.Interface().(T), err)
				return
			}

			for i, a := range dest {
				f = s.Field(i)
				if fieldMap[i] {
					f.Set(reflect.ValueOf(a))
					continue
				}
				f.Set(reflect.ValueOf(a).Elem())
			}
			if !yield(s.Interface().(T), err) {
				return
			}
		}
	}
}
