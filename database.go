package goe

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type conn struct {
	*sql.DB
}

type DB struct {
	conn    conn
	config  Config
	addrMap map[string]any
}

func (db *DB) open(name string, uri string) error {
	if db.conn.DB == nil {
		d, err := sql.Open(name, uri)
		if err == nil {
			db.conn.DB = d
		}
		return err
	}
	return nil
}

type migrateAttribute struct {
	attribute     any
	migrated      bool
	index         bool
	migratedIndex bool
}

type migrateTable struct {
	pk       *pk
	atts     map[string]*migrateAttribute
	migrated bool
}

func (db *DB) Migrate() {
	tables := make(map[string]*migrateTable, 0)
	for _, v := range db.addrMap {
		switch atr := v.(type) {
		case *pk:
			if tables[atr.table] == nil {
				tables[atr.table] = newMigrateTable(db, atr.table)
			}
		}
	}

	for _, t := range tables {
		generateSql(db, t, tables)
		generateIndex(db, t.atts, t.pk.table)
	}
}

func newMigrateTable(db *DB, tableName string) *migrateTable {
	table := new(migrateTable)
	table.atts = make(map[string]*migrateAttribute)
	for _, v := range db.addrMap {
		switch atr := v.(type) {
		case *pk:
			if atr.table == tableName {
				table.pk = atr
				ma := new(migrateAttribute)
				ma.attribute = atr
				table.atts[atr.attributeName] = ma
				for _, fkAny := range atr.fks {
					switch fk := fkAny.(type) {
					case *manyToOne:
						if !fk.hasMany {
							ma := new(migrateAttribute)
							ma.attribute = fk
							table.atts[strings.Split(fk.id, ".")[1]] = ma
						}
					}
				}
			}
		case *att:
			if atr.pk.table == tableName {
				ma := new(migrateAttribute)
				ma.attribute = atr
				//TODO: Add more index
				ma.index = atr.unique
				table.atts[atr.attributeName] = ma
			}
		}
	}

	return table
}

type databaseTable struct {
	columnName   string
	dataType     string
	defaultValue *string
	nullable     bool
}

func generateSql(db *DB, mt *migrateTable, tables map[string]*migrateTable) {
	sqlTableInfos := `SELECT
	column_name, CASE 
	WHEN data_type = 'character varying' 
	THEN CONCAT('varchar','(',character_maximum_length,')')
	WHEN data_type = 'text' THEN 'string'
	WHEN data_type = 'boolean' THEN 'bool'
	WHEN data_type = 'smallint' THEN 'int16'
	WHEN data_type = 'integer' THEN 'int32'
	WHEN data_type = 'bigint' THEN 'int64'
	WHEN data_type = 'real' THEN 'float32'
	WHEN data_type = 'double precision' THEN 'float64'
	ELSE data_type END, 
	column_default, 
	CASE
	WHEN is_nullable = 'YES'
	THEN True
	ELSE False END AS is_nullable
	FROM information_schema.columns WHERE table_name = $1;	
	`
	dataMap := map[string]string{
		"string":  "text",
		"int16":   "smallint",
		"int32":   "integer",
		"int64":   "bigint",
		"float32": "real",
		"float64": "double precision",
	}

	if !mt.migrated {
		rows, err := db.conn.Query(sqlTableInfos, mt.pk.table)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer rows.Close()

		dts := make([]databaseTable, 0)
		dt := databaseTable{}
		for rows.Next() {
			err = rows.Scan(&dt.columnName, &dt.dataType, &dt.defaultValue, &dt.nullable)
			if err != nil {
				fmt.Println(err)
				return
			}
			dts = append(dts, dt)
		}
		if len(dts) > 0 {
			//check fileds
			for i := range dts {
				checkFields(dts[i], mt, tables, dataMap)
			}
			checkNewFields(mt, dataMap)
		} else {
			t := table{}
			fmt.Println("CREATE TABLE " + mt.pk.table)
			for _, attrAny := range mt.atts {
				switch attr := attrAny.attribute.(type) {
				case *pk:
					attr.dataType = checkDataType(attr.dataType, dataMap)
					t.createPk = fmt.Sprintf("%v %v PRIMARY KEY", attr.attributeName, attr.dataType)
				case *att:
					attr.dataType = checkDataType(attr.dataType, dataMap)
					t.createAttrs = append(t.createAttrs, fmt.Sprintf("%v %v %v\n", attr.attributeName, attr.dataType, func(n bool) string {
						if n {
							return "NULL"
						} else {
							return "NOT NULL"
						}
					}(attr.nullable)))
				case *manyToOne:
					pk := tables[attr.targetTable].pk
					pk.dataType = checkDataType(pk.dataType, dataMap)
					fmt.Printf("%v %v %v REFERENCES %v\n", strings.Split(attr.id, ".")[1], pk.dataType, func(n bool) string {
						if n {
							return "NULL"
						} else {
							return "NOT NULL"
						}
					}(attr.nullable), pk.table)
				}
			}
		}
	}
}

type table struct {
	createPk    string
	createAttrs []string
}

type databaseIndex struct {
	attributeName string
	unique        bool
	primaryKey    bool
}

func generateIndex(db *DB, attrs map[string]*migrateAttribute, table string) {
	sql := `SELECT a.attname as attribute, i.indisunique as is_unique, i.indisprimary as is_primary FROM pg_index i
	JOIN pg_attribute a ON i.indexrelid = a.attrelid
	JOIN pg_class ci ON ci.oid = i.indexrelid
	JOIN pg_class c ON c.oid = i.indrelid
	WHERE c.relname = $1;
	`

	rows, err := db.conn.Query(sql, table)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	dis := make([]databaseIndex, 0)
	di := databaseIndex{}
	for rows.Next() {
		err = rows.Scan(&di.attributeName, &di.unique, &di.primaryKey)
		if err != nil {
			fmt.Println(err)
			return
		}
		dis = append(dis, di)
	}
	//check fileds
	for _, di = range dis {
		attrAny := attrs[di.attributeName]
		if attrAny != nil {
			switch attr := attrAny.attribute.(type) {
			case *att:
				fmt.Println("check changes", attr)
			}
		} else {
			fmt.Println(attrAny)
		}
	}
	checkNewIndexs(attrs)

}

func checkNewIndexs(attrs map[string]*migrateAttribute) {
	for _, v := range attrs {
		if v.index && !v.migratedIndex {
			switch attr := v.attribute.(type) {
			case *att:
				fmt.Println(createUniqueIndex(attr.pk.table, attr.attributeName))
			}
		}
	}
}

func createUniqueIndex(table, attribute string) string {
	return fmt.Sprintf("CREATE UNIQUE INDEX ON %v (%v);", table, attribute)
}

func checkFields(databaseTable databaseTable, mt *migrateTable, tables map[string]*migrateTable, dataMap map[string]string) {
	attrAny := mt.atts[databaseTable.columnName]
	if attrAny == nil {
		//fmt.Printf("goe:field '%v'exists on database table but is missed on struct %v\n", databaseTable.columnName, mt.pk.table)
		//prompt a question to drop or set as a rename field
		//attrAny = mt.atts[databaseTable.columnName]
		return
	}
	switch attr := attrAny.attribute.(type) {
	case *pk:
		attr.dataType = checkDataType(attr.dataType, dataMap)
		if databaseTable.dataType != attr.dataType {
			//fmt.Println(alterColumn(attr.table, databaseTable.columnName, attr.dataType, dataMap))
		}
		attrAny.migrated = true
	case *att:
		attr.dataType = checkDataType(attr.dataType, dataMap)
		if databaseTable.dataType != attr.dataType {
			//fmt.Println(alterColumn(attr.pk.table, databaseTable.columnName, attr.dataType, dataMap))
		}
		if databaseTable.nullable != attr.nullable {
			//fmt.Println(nullableColumn(attr.pk.table, attr.attributeName, attr.nullable))
		}
		attrAny.migrated = true
	case *manyToOne:
		attrAny.migrated = true
	}
}

func checkNewFields(mt *migrateTable, dataMap map[string]string) {
	for _, v := range mt.atts {
		if !v.migrated {
			switch attr := v.attribute.(type) {
			case *att:
				attr.dataType = checkDataType(attr.dataType, dataMap)
				//fmt.Println(addColumn(mt.pk.table, attr.attributeName, attr.dataType, attr.nullable))
			}
		}
	}
}

func addColumn(table, column, dataType string, nullable bool) string {
	if nullable {
		return fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v NULL", table, column, dataType)
	}
	return fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v NOT NULL", table, column, dataType)
}

func checkDataType(structDataType string, dataMap map[string]string) string {
	if structDataType == "int8" || structDataType == "uint8" || structDataType == "uint16" {
		structDataType = "int16"
	} else if structDataType == "int" || structDataType == "uint" || structDataType == "uint32" {
		structDataType = "int32"
	} else if structDataType == "uint64" {
		structDataType = "int64"
	}
	if dataMap[structDataType] != "" {
		structDataType = dataMap[structDataType]
	}
	return structDataType
}

func alterColumn(table, column, dataType string, dataMap map[string]string) string {
	if dataMap[dataType] == "" {
		return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v TYPE %v", table, column, dataType)
	}
	return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v TYPE %v", table, column, dataMap[dataType])
}

func nullableColumn(table, columnName string, nullable bool) string {
	if nullable {
		return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v DROP NOT NULL", table, columnName)
	}
	return fmt.Sprintf("ALTER TABLE %v ALTER COLUMN %v SET NOT NULL", table, columnName)
}

func renameColumn(table, oldColumnName, newColumnName string) string {
	return fmt.Sprintf("ALTER TABLE %v RENAME COLUMN %v TO %v", table, oldColumnName, newColumnName)
}

func (db *DB) Select(args ...any) Select {

	stringArgs := getArgs(args...)

	state := createSelectState(db.conn, querySELECT)
	state.addrMap = db.addrMap
	return state.querySelect(stringArgs)
}

func (db *DB) Insert(table any) Insert {
	stringArgs := getArgs(table)

	state := createInsertState(db.conn, queryINSERT)

	return state.queryInsert(stringArgs, db.addrMap)
}

func (db *DB) InsertBetwent(table1 any, table2 any) InsertBetwent {
	stringArgs := getArgs(table1, table2)

	state := createInsertState(db.conn, queryINSERT)

	return state.queryInsertBetwent(stringArgs, db.addrMap)
}

func (db *DB) Update(tables ...any) Update {
	stringArgs := getArgs(tables...)

	state := createUpdateState(db.conn, queryUPDATE)

	return state.queryUpdate(stringArgs, db.addrMap)
}

func (db *DB) UpdateBetwent(table1 any, table2 any) Update {
	stringArgs := getArgs(table1, table2)

	state := createUpdateBetwentState(db.conn, queryUPDATE)

	return state.queryUpdateBetwent(stringArgs, db.addrMap)
}

func (db *DB) Delete(table any) Delete {
	stringArgs := getArgs(table)

	state := createDeleteState(db.conn, queryUPDATE)

	return state.queryDelete(stringArgs, db.addrMap)
}

func (db *DB) Equals(arg any, value any) operator {
	addr := fmt.Sprintf("%p", arg)

	switch atr := db.addrMap[addr].(type) {
	case *att:
		return createComplexOperator(atr.selectName, "=", value, atr.pk)
	case *pk:
		return createComplexOperator(atr.selectName, "=", value, atr)
	}

	return nil
}

func (db *DB) And() operator {
	return simpleOperator{operator: "AND"}
}

func (db *DB) Or() operator {
	return simpleOperator{operator: "OR"}
}

func getArgs(args ...any) []string {
	stringArgs := make([]string, 0)
	for _, v := range args {
		if reflect.ValueOf(v).Kind() == reflect.Ptr {
			if reflect.ValueOf(v).Elem().Kind() == reflect.Struct {
				for i := 0; i < reflect.ValueOf(v).Elem().NumField(); i++ {
					stringArgs = append(stringArgs, fmt.Sprintf("%v", reflect.ValueOf(v).Elem().Field(i).Addr()))
				}
			} else {
				stringArgs = append(stringArgs, fmt.Sprintf("%v", v))
			}
		} else {
			//TODO: Add ptr error
		}
	}
	return stringArgs
}
