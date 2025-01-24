package goe

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/olauro/goe/utils"
)

func AutoMigrate(dbTarget any) error {
	return AutoMigrateContext(context.Background(), dbTarget)
}

func AutoMigrateContext(ctx context.Context, dbTarget any) error {
	db, err := GetGoeDatabase(dbTarget)
	if err != nil {
		return err
	}

	m := migrateFrom(dbTarget, db.Driver)
	if m.Error != nil {
		return m.Error
	}

	sql, err := db.Driver.MigrateContext(ctx, m, db.ConnPool)
	if db.Config.LogQuery {
		log.Println("\n" + sql)
	}
	return err
}

func DropTable(dbTarget any, table string) error {
	db, err := GetGoeDatabase(dbTarget)
	if err != nil {
		return err
	}

	sql, err := db.Driver.DropTable(db.Driver.KeywordHandler(utils.TableNamePattern(table)), db.ConnPool)
	if db.Config.LogQuery {
		log.Println("\n" + sql)
	}
	return err
}

func DropColumn(dbTarget any, table, column string) error {
	db, err := GetGoeDatabase(dbTarget)
	if err != nil {
		return err
	}

	table = db.Driver.KeywordHandler(utils.TableNamePattern(table))
	column = db.Driver.KeywordHandler(utils.ColumnNamePattern(column))

	sql, err := db.Driver.DropColumn(table, column, db.ConnPool)
	if db.Config.LogQuery {
		log.Println("\n" + sql)
	}
	return err
}

func RenameColumn(dbTarget any, table, oldColumn, newColumn string) error {
	db, err := GetGoeDatabase(dbTarget)
	if err != nil {
		return err
	}

	table = db.Driver.KeywordHandler(utils.TableNamePattern(table))
	oldColumn = db.Driver.KeywordHandler(utils.ColumnNamePattern(oldColumn))
	newColumn = db.Driver.KeywordHandler(utils.ColumnNamePattern(newColumn))

	sql, err := db.Driver.RenameColumn(table, oldColumn, newColumn, db.ConnPool)
	if db.Config.LogQuery {
		log.Println("\n" + sql)
	}
	return err
}

func GetGoeDatabase(dbTarget any) (db *DB, err error) {
	dbValueOf := reflect.ValueOf(dbTarget).Elem()
	if dbValueOf.NumField() == 0 {
		return nil, fmt.Errorf("goe: Database %v with no structs", dbValueOf.Type().Name())
	}
	goeDb := AddrMap[uintptr(dbValueOf.Field(0).UnsafePointer())]

	if goeDb == nil {
		return nil, fmt.Errorf("goe: Database %v with no structs", dbValueOf.Type().Name())
	}

	return goeDb.GetDb(), nil
}
