package goe

import (
	"context"
	"log"

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

	sql, err := db.Driver.MigrateContext(ctx, m, db.SqlDB)
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

	sql, err := db.Driver.DropTable(db.Driver.KeywordHandler(utils.TableNamePattern(table)), db.SqlDB)
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

	sql, err := db.Driver.DropColumn(table, column, db.SqlDB)
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

	sql, err := db.Driver.RenameColumn(table, oldColumn, newColumn, db.SqlDB)
	if db.Config.LogQuery {
		log.Println("\n" + sql)
	}
	return err
}
