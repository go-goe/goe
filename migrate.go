package goe

import (
	"context"

	"github.com/go-goe/goe/utils"
)

func AutoMigrate(dbTarget any) error {
	return AutoMigrateContext(context.Background(), dbTarget)
}

func AutoMigrateContext(ctx context.Context, dbTarget any) error {
	db := getDatabase(dbTarget)

	m := migrateFrom(dbTarget, db.driver)
	if m.Error != nil {
		return m.Error
	}

	return db.driver.MigrateContext(ctx, m)
}

func DropTable(dbTarget any, schema, table string) error {
	db := getDatabase(dbTarget)

	schema = db.driver.KeywordHandler(utils.ColumnNamePattern(schema))
	table = db.driver.KeywordHandler(utils.TableNamePattern(table))
	return db.driver.DropTable(schema, table)
}

func DropColumn(dbTarget any, schema, table, column string) error {
	db := getDatabase(dbTarget)

	schema = db.driver.KeywordHandler(utils.ColumnNamePattern(schema))
	table = db.driver.KeywordHandler(utils.TableNamePattern(table))
	column = db.driver.KeywordHandler(utils.ColumnNamePattern(column))

	return db.driver.DropColumn(schema, table, column)
}

func RenameColumn(dbTarget any, schema, table, oldColumn, newColumn string) error {
	db := getDatabase(dbTarget)

	schema = db.driver.KeywordHandler(utils.ColumnNamePattern(schema))
	table = db.driver.KeywordHandler(utils.TableNamePattern(table))
	oldColumn = db.driver.KeywordHandler(utils.ColumnNamePattern(oldColumn))
	newColumn = db.driver.KeywordHandler(utils.ColumnNamePattern(newColumn))

	return db.driver.RenameColumn(schema, table, oldColumn, newColumn)
}
