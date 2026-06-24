package goe

import (
	"context"

	"github.com/go-goe/goe/utils"
)

type migrate struct {
	db       *DB
	dbTarget any
}

type migrateSchema struct {
	migrate
	schema string
}

type migrateTable struct {
	migrateSchema
	table string
}

func Migrate(dbTarget any) migrate {
	return migrate{db: getDatabase(dbTarget), dbTarget: dbTarget}
}

func (m migrate) AutoMigrate() error {
	return m.AutoMigrateContext(context.Background())
}

func (m migrate) OnSchema(schema string) migrateSchema {
	return migrateSchema{m, schema}
}

func (m migrate) OnTable(table string) migrateTable {
	return migrateTable{migrateSchema{migrate: m}, table}
}

func (ms migrateSchema) OnTable(table string) migrateTable {
	return migrateTable{ms, table}
}

func (m migrate) AutoMigrateContext(ctx context.Context) error {
	data, err := MigrateFrom(m.dbTarget, m.db.driver)
	if err != nil {
		return err
	}
	return m.db.driver.MigrateContext(ctx, data)
}

func (mt migrateTable) DropTable() error {
	return mt.db.driver.DropTable(
		mt.db.driver.KeywordHandler(utils.ColumnNamePattern(mt.schema)),
		mt.db.driver.KeywordHandler(utils.TableNamePattern(mt.table)))
}

func (mt migrateTable) RenameTable(newName string) error {
	return mt.db.driver.RenameTable(
		mt.db.driver.KeywordHandler(utils.ColumnNamePattern(mt.schema)),
		mt.db.driver.KeywordHandler(utils.TableNamePattern(mt.table)),
		mt.db.driver.KeywordHandler(utils.TableNamePattern(newName)))
}

func (mt migrateTable) DropColumn(column string) error {
	return mt.db.driver.DropColumn(
		mt.db.driver.KeywordHandler(utils.ColumnNamePattern(mt.schema)),
		mt.db.driver.KeywordHandler(utils.TableNamePattern(mt.table)),
		mt.db.driver.KeywordHandler(utils.ColumnNamePattern(column)))
}

func (mt migrateTable) RenameColumn(column, newName string) error {
	return mt.db.driver.RenameColumn(
		mt.db.driver.KeywordHandler(utils.ColumnNamePattern(mt.schema)),
		mt.db.driver.KeywordHandler(utils.TableNamePattern(mt.table)),
		mt.db.driver.KeywordHandler(utils.ColumnNamePattern(column)),
		mt.db.driver.KeywordHandler(utils.ColumnNamePattern(newName)))
}
