package goe

import (
	"context"
	"database/sql"
	"iter"
	"reflect"
)

func handlerValues(conn Connection, sqlQuery string, args []any, ctx context.Context) error {
	_, err := conn.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return err
	}
	return nil
}

func handlerValuesReturning(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string, ctx context.Context) error {
	row := conn.QueryRowContext(ctx, sqlQuery, args...)

	err := row.Scan(value.FieldByName(idName).Addr().Interface())
	if err != nil {
		return err
	}
	return nil
}

func handlerValuesReturningBatch(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string, ctx context.Context) error {
	rows, err := conn.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		err = rows.Scan(value.Index(i).FieldByName(idName).Addr().Interface())
		if err != nil {
			return err
		}
		i++
	}
	return nil
}

func handlerResult[T any](conn Connection, sqlQuery string, args []any, numFields int, ctx context.Context) iter.Seq2[T, error] {
	rows, err := conn.QueryContext(ctx, sqlQuery, args...)

	var v T
	if err != nil {
		return func(yield func(T, error) bool) {
			yield(v, err)
		}
	}

	value := reflect.TypeOf(v)
	dest := make([]any, numFields)
	for i := range dest {
		dest[i] = reflect.New(value.Field(i).Type).Interface()
	}

	return mapStructQuery[T](rows, dest, value)
}

func mapStructQuery[T any](rows *sql.Rows, dest []any, value reflect.Type) iter.Seq2[T, error] {
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
