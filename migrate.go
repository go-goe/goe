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

	m := migrateFrom(dbTarget, db.Driver)
	if m.Error != nil {
		return m.Error
	}

	return db.Driver.MigrateContext(ctx, m)
}

func DropTable(dbTarget any, table string) error {
	db := getDatabase(dbTarget)

	return db.Driver.DropTable(db.Driver.KeywordHandler(utils.TableNamePattern(table)))
}

func DropColumn(dbTarget any, table, column string) error {
	db := getDatabase(dbTarget)

	table = db.Driver.KeywordHandler(utils.TableNamePattern(table))
	column = db.Driver.KeywordHandler(utils.ColumnNamePattern(column))

	return db.Driver.DropColumn(table, column)
}

func RenameColumn(dbTarget any, table, oldColumn, newColumn string) error {
	db := getDatabase(dbTarget)

	table = db.Driver.KeywordHandler(utils.TableNamePattern(table))
	oldColumn = db.Driver.KeywordHandler(utils.ColumnNamePattern(oldColumn))
	newColumn = db.Driver.KeywordHandler(utils.ColumnNamePattern(newColumn))

	return db.Driver.RenameColumn(table, oldColumn, newColumn)
}
