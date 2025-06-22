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

func DropTable(dbTarget any, scheme, table string) error {
	db := getDatabase(dbTarget)

	scheme = db.driver.KeywordHandler(utils.ColumnNamePattern(scheme))
	table = db.driver.KeywordHandler(utils.TableNamePattern(table))
	return db.driver.DropTable(scheme, table)
}

func DropColumn(dbTarget any, scheme, table, column string) error {
	db := getDatabase(dbTarget)

	scheme = db.driver.KeywordHandler(utils.ColumnNamePattern(scheme))
	table = db.driver.KeywordHandler(utils.TableNamePattern(table))
	column = db.driver.KeywordHandler(utils.ColumnNamePattern(column))

	return db.driver.DropColumn(scheme, table, column)
}

func RenameColumn(dbTarget any, scheme, table, oldColumn, newColumn string) error {
	db := getDatabase(dbTarget)

	scheme = db.driver.KeywordHandler(utils.ColumnNamePattern(scheme))
	table = db.driver.KeywordHandler(utils.TableNamePattern(table))
	oldColumn = db.driver.KeywordHandler(utils.ColumnNamePattern(oldColumn))
	newColumn = db.driver.KeywordHandler(utils.ColumnNamePattern(newColumn))

	return db.driver.RenameColumn(scheme, table, oldColumn, newColumn)
}
