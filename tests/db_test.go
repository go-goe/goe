package tests_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/go-goe/goe"
	"github.com/go-goe/postgres"
	"github.com/go-goe/sqlite"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Animal struct {
	Name        string `goe:"index"`
	IdHabitat   *uuid.UUID
	IdInfo      *[]byte
	Id          int
	AnimalFoods []AnimalFood
}

type AnimalFood struct {
	IdAnimal int       `goe:"pk"`
	IdFood   uuid.UUID `goe:"pk"`
}

type Food struct {
	Id          uuid.UUID
	Name        string
	AnimalFoods []AnimalFood
}

type Habitat struct {
	Id          uuid.UUID
	Name        string `goe:"type:varchar(50)"`
	IdWeather   int
	NameWeather string
	Animals     []Animal
}

type Weather struct {
	Id       int `goe:"pk"`
	Name     string
	Habitats []Habitat
}

type Info struct {
	Id         []byte
	Name       string `goe:"index(unique n:idx_name_status);index"`
	NameStatus string `goe:"index(unique n:idx_name_status)"`
	IdStatus   int
}

type Status struct {
	Id   int
	Name string
}

type User struct {
	Id        int
	Name      string
	Email     string `goe:"unique"`
	UserRoles []UserRole
}

type UserRole struct {
	Id      int
	UserId  int
	RoleId  int
	EndDate *time.Time
}

type Role struct {
	Id        int
	Name      string
	UserRoles []UserRole
}

type Flag struct {
	Id         uuid.UUID
	Name       string
	Float32    float32
	Float64    float64
	Today      time.Time
	Int        int
	Int8       int8
	Int16      int16
	Int32      int32
	Int64      int64
	Uint       uint
	Uint8      uint8
	Uint16     uint16
	Uint32     uint32 `goe:"default:32"`
	Uint64     uint64
	Bool       bool
	Price      decimal.Decimal `goe:"type:decimal(10,4)"`
	Byte       []byte
	NullId     sql.Null[uuid.UUID] `goe:"type:uuid"`
	NullString sql.NullString      `goe:"type:varchar(100)"`
}

type Person struct {
	Id   int
	Name string
	Jobs []JobTitle
}

type PersonJobTitle struct {
	PersonId   int `goe:"pk"`
	IdJobTitle int `goe:"pk"`
	CreatedAt  time.Time
}

type JobTitle struct {
	Name    string
	Id      int
	Persons []Person
}

type Exam struct {
	Id      int
	Score   float32
	Minimum float32
}

type Select struct {
	Id   int
	Name string
}

type Page struct {
	Id     int
	Number int
	PageId *int
}

type FlagSchema struct {
	Flag *Flag
}

type Authentication struct {
	User     *User
	UserRole *UserRole
	Role     *Role
}

type FoodHabitatSchema struct {
	Food    *Food
	Habitat *Habitat
}

type Drop struct {
	Id   int
	Name string
}

type DropSchema struct {
	Drop *Drop
}

type Database struct {
	Animal     *Animal
	AnimalFood *AnimalFood
	*FoodHabitatSchema
	Info            *Info
	Status          *Status
	Weather         *Weather
	*Authentication `goe:"schema"`
	*FlagSchema
	Person         *Person
	PersonJobTitle *PersonJobTitle
	JobTitle       *JobTitle
	Exam           *Exam
	Select         *Select
	Page           *Page
	*DropSchema
	*goe.DB
}

var db *Database

var mapDriver = map[string]func() (*Database, error){
	"PostgreSQL": SetupPostgres,
	"SQLite":     SetupSqlite,
}

func Setup() (*Database, error) {
	if db != nil {
		return db, nil
	}
	var err error
	db, err = mapDriver[os.Getenv("GOE_DRIVER")]()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func SetupPostgres() (*Database, error) {
	var err error
	db, err := goe.Open[Database](postgres.Open("user=postgres password=postgres host=localhost port=5432 database=postgres", postgres.Config{
		//DatabaseConfig: goe.DatabaseConfig{Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil))},
	}))
	if err != nil {
		return nil, err
	}
	err = goe.Migrate(db).AutoMigrate()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func SetupSqlite() (*Database, error) {
	var err error
	db, err := goe.Open[Database](sqlite.Open(filepath.Join(os.TempDir(), "goe.db"), sqlite.Config{
		//DatabaseConfig: goe.DatabaseConfig{Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil))},
	}))
	if err != nil {
		return nil, err
	}
	err = goe.Migrate(db).AutoMigrate()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestConnection(t *testing.T) {
	_, err := Setup()
	if err != nil {
		t.Fatalf("Expected Connection, got error %v", err)
	}
}

func TestTx(t *testing.T) {
	db, err := Setup()
	if err != nil {
		t.Fatalf("Expected setup, got error %v", err)
	}

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Tx_Context_Cancel",
			testCase: func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				_, err := db.NewTransactionContext(ctx, sql.LevelSerializable)
				if !errors.Is(err, context.Canceled) {
					t.Errorf("Expected context.Canceled, got : %v", err)
				}
			},
		},
		{
			desc: "Tx_Context_Timeout",
			testCase: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				defer cancel()

				_, err := db.NewTransactionContext(ctx, sql.LevelSerializable)
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("Expected context.DeadlineExceeded, got : %v", err)
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}

func TestRace(t *testing.T) {
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			raceDb, _ := mapDriver[os.Getenv("GOE_DRIVER")]()
			goe.Close(raceDb)
		}()
	}
	wg.Wait()
}

func TestMigrate(t *testing.T) {
	db, err := Setup()
	if err != nil {
		t.Fatalf("Expected a connection, got error %v", err)
	}
	err = goe.Migrate(db).OnTable("Select").RenameColumn("Name", "NewName")
	if err != nil {
		t.Fatalf("Expected rename column, got error %v", err)
	}

	err = goe.Migrate(db).OnTable("Select").DropColumn("NewName")
	if err != nil {
		t.Fatalf("Expected drop column, got error %v", err)
	}

	err = goe.Migrate(db).OnTable("Select").RenameTable("NewSelect")
	if err != nil {
		t.Fatalf("Expected rename table, got error %v", err)
	}

	err = goe.Migrate(db).OnTable("NewSelect").DropTable()
	if err != nil {
		t.Fatalf("Expected drop table NewSelect, got error %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = goe.Migrate(db).AutoMigrateContext(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected context.Canceled, got %v", err)
	}

	err = goe.Migrate(db).OnSchema("DropSchema").OnTable("Drop").RenameColumn("Name", "NewName")
	if err != nil {
		t.Fatalf("Expected rename column, got error %v", err)
	}

	err = goe.Migrate(db).OnSchema("DropSchema").OnTable("Drop").DropColumn("NewName")
	if err != nil {
		t.Fatalf("Expected drop column, got error %v", err)
	}

	err = goe.Migrate(db).OnSchema("DropSchema").OnTable("Drop").DropTable()
	if err != nil {
		t.Fatalf("Expected drop table DropSchema.Drop, got error %v", err)
	}
}
