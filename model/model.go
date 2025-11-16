package model

import (
	"context"
	"time"

	"github.com/go-goe/goe/enum"
)

type Attribute struct {
	Table         string
	Name          string
	AggregateType enum.AggregateType
	FunctionType  enum.FunctionType
}

type JoinArgument struct {
	Table string
	Name  string
}

type Join struct {
	Table          Table
	FirstArgument  JoinArgument
	JoinOperation  enum.JoinType
	SecondArgument JoinArgument
}

type Where struct {
	Type           enum.WhereType
	Attribute      Attribute
	Operator       enum.OperatorType
	AttributeValue Attribute
	SizeIn         uint
	QueryIn        *Query
}

type OrderBy struct {
	Desc      bool
	Attribute Attribute
}

type GroupBy struct {
	Attribute Attribute
}

type Table struct {
	Schema *string
	Name   string
}

func (t Table) String() string {
	if t.Schema != nil {
		return *t.Schema + "." + t.Name
	}
	return t.Name
}

type Query struct {
	Type       enum.QueryType
	Attributes []Attribute
	Tables     []Table

	Joins   []Join    //Select
	Limit   int       //Select
	Offset  int       //Select
	OrderBy []OrderBy //Select
	GroupBy []GroupBy //Select

	WhereOperations []Where //Select, Update and Delete
	WhereIndex      int     //Start of where position arguments $1, $2...
	Arguments       []any

	ReturningID    *Attribute //Insert
	BatchSizeQuery int        //Insert
	SizeArguments  int        //Insert

	RawSql string
	Header QueryHeader
}

type QueryHeader struct {
	Err           error
	ModelBuild    time.Duration
	QueryDuration time.Duration
}

type Operation struct {
	Type                enum.WhereType
	Arg                 any
	Value               ValueOperation
	Operator            enum.OperatorType
	Attribute           string
	Table               Table
	TableId             int
	Function            enum.FunctionType
	AttributeValue      string
	AttributeValueTable Table
	AttributeTableId    int
	FirstOperation      *Operation
	SecondOperation     *Operation
}

type Set struct {
	Attribute any
	Value     any
}

type Body struct {
	Table string
	Name  string
}

type Migrator struct {
	Tables  map[string]*TableMigrate
	Schemas []string
	Error   error
}

type TableMigrate struct {
	Name         string
	EscapingName string
	Schema       *string
	Migrated     bool
	PrimaryKeys  []PrimaryKeyMigrate
	Attributes   []AttributeMigrate
	ManyToOnes   []ManyToOneMigrate
	OneToOnes    []OneToOneMigrate
	Indexes      []IndexMigrate
}

// Returns the table and the schema.
func (t TableMigrate) EscapingTableName() string {
	if t.Schema != nil {
		return *t.Schema + "." + t.EscapingName
	}
	return t.EscapingName
}

type IndexMigrate struct {
	Name         string
	EscapingName string
	Unique       bool
	Attributes   []AttributeMigrate
}

type PrimaryKeyMigrate struct {
	AutoIncrement bool
	AttributeMigrate
}

type AttributeMigrate struct {
	Nullable     bool
	Name         string
	EscapingName string
	DataType     string
	Default      string
}

type OneToOneMigrate struct {
	AttributeMigrate
	TargetTable          string
	TargetColumn         string
	EscapingTargetTable  string
	EscapingTargetColumn string
	TargetSchema         *string
}

// Returns the target table and the schema.
func (o OneToOneMigrate) EscapingTargetTableName() string {
	if o.TargetSchema != nil {
		return *o.TargetSchema + "." + o.EscapingTargetTable
	}
	return o.EscapingTargetTable
}

type ManyToOneMigrate struct {
	AttributeMigrate
	TargetTable          string
	TargetColumn         string
	EscapingTargetTable  string
	EscapingTargetColumn string
	TargetSchema         *string
}

// Returns the target table and the schema.
func (m ManyToOneMigrate) EscapingTargetTableName() string {
	if m.TargetSchema != nil {
		return *m.TargetSchema + "." + m.EscapingTargetTable
	}
	return m.EscapingTargetTable
}

// Database config used by all GOE drivers
type DatabaseConfig struct {
	Logger           Logger
	IncludeArguments bool          // include all arguments used on query
	QueryThreshold   time.Duration // query threshold to warning on slow queries
	databaseName     string
	errorTranslator  func(err error) error
	schemas          []string
	initCallback     func() error
}

func (c DatabaseConfig) ErrorHandler(ctx context.Context, err error) error {
	if c.Logger != nil {
		c.Logger.ErrorContext(ctx, "error", "database", c.databaseName, "err", err)
	}
	return err
}

func (c DatabaseConfig) ErrorQueryHandler(ctx context.Context, query Query) error {
	query.Header.Err = c.errorTranslator(query.Header.Err)
	if c.Logger == nil {
		return query.Header.Err
	}
	logs := make([]any, 0)
	logs = append(logs, "database", c.databaseName)
	logs = append(logs, "sql", query.RawSql)
	if c.IncludeArguments {
		logs = append(logs, "arguments", query.Arguments)
	}
	logs = append(logs, "err", query.Header.Err)

	c.Logger.ErrorContext(ctx, "error", logs...)
	return query.Header.Err
}

func (c DatabaseConfig) InfoHandler(ctx context.Context, query Query) {
	if c.Logger == nil {
		return
	}
	qr := query.Header.QueryDuration + query.Header.ModelBuild

	logs := make([]any, 0)
	logs = append(logs, "database", c.databaseName)
	logs = append(logs, "query_duration", qr.String())
	logs = append(logs, "sql", query.RawSql)
	if c.IncludeArguments {
		logs = append(logs, "arguments", query.Arguments)
	}

	if c.QueryThreshold != 0 && qr > c.QueryThreshold {
		c.Logger.WarnContext(ctx, "query_threshold", logs...)
		return
	}

	c.Logger.InfoContext(ctx, "query_runned", logs...)
}

func (c DatabaseConfig) Schemas() []string {
	return c.schemas
}

func (c *DatabaseConfig) SetSchemas(s []string) {
	c.schemas = s
}

func (c *DatabaseConfig) SetInitCallback(f func() error) {
	c.initCallback = f
}

func (c DatabaseConfig) InitCallback() func() error {
	return c.initCallback
}

func (c *DatabaseConfig) Init(driverName string, errorTranslator func(err error) error) {
	c.schemas = nil
	c.initCallback = nil
	c.databaseName = driverName
	c.errorTranslator = errorTranslator
}
