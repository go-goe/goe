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

type handlerNullable struct {
	entityField     reflect.Value
	noNullableValue any
}

func handlerResult[T any](ctx context.Context, conn model.Connection, query model.Query, numFields int, dbConfig *model.DatabaseConfig) iter.Seq2[T, error] {
	var rows model.Rows
	rows, query.Header.Err = wrapperQuery(ctx, conn, &query)

	var entity T
	if query.Header.Err != nil {
		return func(yield func(T, error) bool) {
			yield(entity, dbConfig.ErrorQueryHandler(ctx, query))
		}
	}
	dbConfig.InfoHandler(ctx, query)

	dest := make([]any, numFields)
	canBeNullable := make([]handlerNullable, numFields)
	value := reflect.ValueOf(&entity).Elem()
	var noNullableValue any
	for i := range dest {
		if value.Field(i).Kind() != reflect.Pointer {
			noNullableValue = reflect.New(value.Field(i).Addr().Type()).Interface()

			canBeNullable[i] = handlerNullable{
				entityField:     value.Field(i),
				noNullableValue: noNullableValue,
			}

			dest[i] = noNullableValue
			continue
		}
		dest[i] = value.Field(i).Addr().Interface()
	}

	return func(yield func(T, error) bool) {
		defer rows.Close()

		for rows.Next() {
			query.Header.Err = rows.Scan(dest...)
			if query.Header.Err != nil {
				//TODO: add infos about row
				yield(entity, dbConfig.ErrorQueryHandler(ctx, query))
				return
			}

			for _, nullable := range canBeNullable {
				valueOf := reflect.ValueOf(nullable.noNullableValue)
				if valueOf.IsValid() && !valueOf.IsNil() {
					valueOf = valueOf.Elem()
					if valueOf.IsValid() && !valueOf.IsNil() {
						nullable.entityField.Set(valueOf.Elem())
					}
				}
			}
			if !yield(entity, nil) {
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
