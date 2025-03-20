package goe

import (
	"context"

	"github.com/olauro/goe/utils"
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

func DropTable(dbTarget any, table string) error {
	db := getDatabase(dbTarget)

	return db.driver.DropTable(db.driver.KeywordHandler(utils.TableNamePattern(table)))
}

func DropColumn(dbTarget any, table, column string) error {
	db := getDatabase(dbTarget)

	table = db.driver.KeywordHandler(utils.TableNamePattern(table))
	column = db.driver.KeywordHandler(utils.ColumnNamePattern(column))

	return db.driver.DropColumn(table, column)
}

func RenameColumn(dbTarget any, table, oldColumn, newColumn string) error {
	db := getDatabase(dbTarget)

	table = db.driver.KeywordHandler(utils.TableNamePattern(table))
	oldColumn = db.driver.KeywordHandler(utils.ColumnNamePattern(oldColumn))
	newColumn = db.driver.KeywordHandler(utils.ColumnNamePattern(newColumn))

	return db.driver.RenameColumn(table, oldColumn, newColumn)
}
