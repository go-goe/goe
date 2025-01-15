package migrate

import (
	"context"
	"log"

	"github.com/olauro/goe"
	"github.com/olauro/goe/utils"
)

func AutoMigrate(db *goe.DB, m *goe.Migrator) error {
	return AutoMigrateContext(context.Background(), db, m)
}

func AutoMigrateContext(ctx context.Context, db *goe.DB, m *goe.Migrator) error {
	if m.Error != nil {
		return m.Error
	}
	sql, err := db.Driver.MigrateContext(ctx, m, db.ConnPool)
	if db.Config.LogQuery {
		log.Println("\n" + sql)
	}
	return err
}

func DropTable(db *goe.DB, table string) error {
	sql, err := db.Driver.DropTable(db.Driver.KeywordHandler(utils.TableNamePattern(table)), db.ConnPool)
	if db.Config.LogQuery {
		log.Println("\n" + sql)
	}
	return err
}

func DropColumn(db *goe.DB, table, column string) error {
	table = db.Driver.KeywordHandler(utils.TableNamePattern(table))
	column = db.Driver.KeywordHandler(utils.ColumnNamePattern(column))

	sql, err := db.Driver.DropColumn(table, column, db.ConnPool)
	if db.Config.LogQuery {
		log.Println("\n" + sql)
	}
	return err
}

func RenameColumn(db *goe.DB, table, oldColumn, newColumn string) error {
	table = db.Driver.KeywordHandler(utils.TableNamePattern(table))
	oldColumn = db.Driver.KeywordHandler(utils.ColumnNamePattern(oldColumn))
	newColumn = db.Driver.KeywordHandler(utils.ColumnNamePattern(newColumn))

	sql, err := db.Driver.RenameColumn(table, oldColumn, newColumn, db.ConnPool)
	if db.Config.LogQuery {
		log.Println("\n" + sql)
	}
	return err
}
