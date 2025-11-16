package goe

import (
	"context"
	"iter"
	"reflect"
	"time"

	"github.com/go-goe/goe/model"
)

func handlerValues(ctx context.Context, conn model.Connection, query model.Query, dbConfig *model.DatabaseConfig) error {
	query.Header.Err = wrapperExec(ctx, conn, &query)
	if query.Header.Err != nil {
		return dbConfig.ErrorQueryHandler(ctx, query)
	}
	dbConfig.InfoHandler(ctx, query)
	return nil
}

func handlerValuesReturning(ctx context.Context, conn model.Connection, query model.Query, value reflect.Value, pkFieldId int, dbConfig *model.DatabaseConfig) error {
	row := wrapperQueryRow(ctx, conn, &query)

	query.Header.Err = row.Scan(value.Field(pkFieldId).Addr().Interface())
	if query.Header.Err != nil {
		return dbConfig.ErrorQueryHandler(ctx, query)
	}
	dbConfig.InfoHandler(ctx, query)
	return nil
}

func handlerValuesReturningBatch(ctx context.Context, conn model.Connection, query model.Query, value reflect.Value, pkFieldId int, dbConfig *model.DatabaseConfig) error {
	var rows model.Rows
	rows, query.Header.Err = wrapperQuery(ctx, conn, &query)

	if query.Header.Err != nil {
		return dbConfig.ErrorQueryHandler(ctx, query)
	}
	defer rows.Close()
	dbConfig.InfoHandler(ctx, query)

	i := 0
	for rows.Next() {
		query.Header.Err = rows.Scan(value.Index(i).Field(pkFieldId).Addr().Interface())
		if query.Header.Err != nil {
			//TODO: add infos about row
			return dbConfig.ErrorQueryHandler(ctx, query)
		}
		i++
	}
	return nil
}

func handlerResult[T any](ctx context.Context, conn model.Connection, query model.Query, numFields int, dbConfig *model.DatabaseConfig) iter.Seq2[T, error] {
	var rows model.Rows
	rows, query.Header.Err = wrapperQuery(ctx, conn, &query)

	var v T
	if query.Header.Err != nil {
		return func(yield func(T, error) bool) {
			yield(v, dbConfig.ErrorQueryHandler(ctx, query))
		}
	}
	dbConfig.InfoHandler(ctx, query)

	value := reflect.TypeOf(v)
	dest := make([]any, numFields)

	for i := range dest {
		dest[i] = reflect.New(value.Field(i).Type).Interface()
	}

	return mapStructQuery[T](ctx, rows, dest, value, dbConfig, query)
}

func mapStructQuery[T any](ctx context.Context, rows model.Rows, dest []any, value reflect.Type, dbConfig *model.DatabaseConfig, query model.Query) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var (
			s, f reflect.Value
		)
		defer rows.Close()
		s = reflect.New(value).Elem()

		for rows.Next() {
			query.Header.Err = rows.Scan(dest...)

			if query.Header.Err != nil {
				//TODO: add infos about row
				yield(s.Interface().(T), dbConfig.ErrorQueryHandler(ctx, query))
				return
			}

			for i, a := range dest {
				f = s.Field(i)
				f.Set(reflect.ValueOf(a).Elem())
			}
			if !yield(s.Interface().(T), nil) {
				return
			}
		}
	}
}

func wrapperQuery(ctx context.Context, conn model.Connection, query *model.Query) (model.Rows, error) {
	queryStart := time.Now()
	defer func() { query.Header.QueryDuration = time.Since(queryStart) }()
	return conn.QueryContext(ctx, query)
}

func wrapperQueryRow(ctx context.Context, conn model.Connection, query *model.Query) model.Row {
	queryStart := time.Now()
	defer func() { query.Header.QueryDuration = time.Since(queryStart) }()
	return conn.QueryRowContext(ctx, query)
}

func wrapperExec(ctx context.Context, conn model.Connection, query *model.Query) error {
	queryStart := time.Now()
	defer func() { query.Header.QueryDuration = time.Since(queryStart) }()
	return conn.ExecContext(ctx, query)
}
