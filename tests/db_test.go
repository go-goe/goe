package tests_test

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
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

/* -------------------------- Models -------------------------- */

type FlagModel struct {
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
	Uint32     uint32
	Uint64     uint64
	Bool       bool
	Price      decimal.Decimal
	Byte       []byte
	NullId     sql.Null[uuid.UUID]
	NullString sql.NullString
}

type PersonJobTitleModel struct {
	PersonId   int
	JobTitleId int
	CreatedAt  time.Time
}

type PersonModel struct {
	Id   int
	Name string
	Jobs []JobTitle
}

type JobTitleModel struct {
	Name    string
	Id      int
	Persons []Person
}

type InfoModel struct {
	Id         []byte
	Name       string
	NameStatus string
	StatusId   int
}

type AnimalModel struct {
	Name      string
	HabitatId *uuid.UUID
	InfoId    *[]byte
	Id        int
}

type AnimalFoodModel struct {
	AnimalId int
	FoodId   uuid.UUID
}

/* -------------------------- Entitys -------------------------- */

type Animal struct {
	Name        goe.Type[string] `goe:"index"`
	HabitatId   goe.TypeNull[uuid.UUID]
	InfoId      goe.TypeNull[[]byte]
	Id          goe.Type[int]
	AnimalFoods []AnimalFood
	goe.EntityMap[Animal, AnimalModel]
}

type AnimalFood struct {
	AnimalId goe.Type[int]       `goe:"pk"`
	FoodId   goe.Type[uuid.UUID] `goe:"pk"`
	goe.EntityMap[AnimalFood, AnimalFoodModel]
}

type Food struct {
	Id          goe.Type[uuid.UUID]
	Name        goe.Type[string]
	AnimalFoods []AnimalFood
	goe.EntityMap[Food, FoodModel]
}

type FoodModel struct {
	Id          uuid.UUID
	Name        string
	AnimalFoods []AnimalFood
}

type Habitat struct {
	Id          goe.Type[uuid.UUID]
	Name        goe.Type[string] `goe:"type:varchar(50)"`
	WeatherId   goe.Type[int]
	NameWeather goe.Type[string]
	Animals     []Animal
	goe.EntityMap[Habitat, HabitatModel]
}

type HabitatModel struct {
	Id          uuid.UUID
	Name        string
	WeatherId   int
	NameWeather string
	Animals     []Animal
}

type Weather struct {
	Id       goe.Type[int] `goe:"pk"`
	Name     goe.Type[string]
	Habitats []Habitat
	goe.EntityMap[Weather, WeatherModel]
}

type WeatherModel struct {
	Id       int
	Name     string
	Habitats []Habitat
}

type Info struct {
	Id         goe.Type[[]byte]
	Name       goe.Type[string] `goe:"index(unique n:idx_name_status);index"`
	NameStatus goe.Type[string] `goe:"index(unique n:idx_name_status)"`
	StatusId   goe.Type[int]
	goe.EntityMap[Info, InfoModel]
}

type Status struct {
	Id   goe.Type[int]
	Name goe.Type[string]
	goe.EntityMap[Status, StatusModel]
}

type StatusModel struct {
	Id   int
	Name string
}

type User struct {
	Id        goe.Type[int]
	Name      goe.Type[string] `goe:"index(n:idx_name_lower f:lower)"`
	Email     goe.Type[string] `goe:"unique"`
	UserRoles []UserRole
	goe.EntityMap[User, UserModel]
}

type UserModel struct {
	Id        int
	Name      string
	Email     string
	UserRoles []UserRole
}

type UserRole struct {
	Id      goe.Type[int]
	UserId  goe.Type[int]
	RoleId  goe.Type[int]
	EndDate goe.TypeNull[time.Time]
	goe.EntityMap[UserRole, UserRoleModel]
}

type UserRoleModel struct {
	Id      int
	UserId  int
	RoleId  int
	EndDate *time.Time
}

type Role struct {
	Id        goe.Type[int]
	Name      goe.Type[string]
	UserRoles []UserRole
	goe.EntityMap[Role, RoleModel]
}

type RoleModel struct {
	Id        int
	Name      string
	UserRoles []UserRole
}

type Flag struct {
	Id         goe.Type[uuid.UUID]
	Name       goe.Type[string]
	Float32    goe.Type[float32]
	Float64    goe.Type[float64]
	Today      goe.Type[time.Time]
	Int        goe.Type[int]
	Int8       goe.Type[int8]
	Int16      goe.Type[int16]
	Int32      goe.Type[int32]
	Int64      goe.Type[int64]
	Uint       goe.Type[uint]
	Uint8      goe.Type[uint8]
	Uint16     goe.Type[uint16]
	Uint32     goe.Type[uint32] `goe:"default:32"`
	Uint64     goe.Type[uint64]
	Bool       goe.Type[bool]
	Price      goe.Type[decimal.Decimal] `goe:"type:decimal(10,4)"`
	Byte       goe.Type[[]byte]
	NullId     goe.TypeNull[sql.Null[uuid.UUID]] `goe:"type:uuid"`
	NullString goe.TypeNull[sql.NullString]      `goe:"type:varchar(100)"`
	goe.EntityMap[Flag, FlagModel]
}

type Person struct {
	Id   goe.Type[int]
	Name goe.Type[string]
	Jobs []JobTitle
	goe.EntityMap[Person, PersonModel]
}

type PersonJobTitle struct {
	PersonId   goe.Type[int] `goe:"pk"`
	JobTitleId goe.Type[int] `goe:"pk"`
	CreatedAt  goe.Type[time.Time]
	goe.EntityMap[PersonJobTitle, PersonJobTitleModel]
}

type JobTitle struct {
	Name    goe.Type[string]
	Id      goe.Type[int]
	Persons []Person
	goe.EntityMap[JobTitle, JobTitleModel]
}

type Exam struct {
	Id      goe.Type[int]
	Score   goe.Type[float32]
	Minimum goe.Type[float32]
	goe.EntityMap[Exam, ExamModel]
}

type ExamModel struct {
	Id      int
	Score   float32
	Minimum float32
}

type Insert struct {
	Id   goe.Type[int]
	Name goe.Type[string]
	goe.EntityMap[Insert, InsertModel]
}

type InsertModel struct {
	Id   int
	Name string
}

type Page struct {
	ID         goe.Type[int]
	Number     goe.Type[int]
	PageIDNext goe.TypeNull[int]
	PageIDPrev goe.TypeNull[int]
	goe.EntityMap[Page, PageModel]
}

type PageModel struct {
	ID         int
	Number     int
	PageIDNext *int
	PageIDPrev *int
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
	Id   goe.Type[int]
	Name goe.Type[string]
	goe.EntityMap[Drop, DropModel]
}

type DropModel struct {
	Id   int
	Name string
}

type DropSchema struct {
	Drop *Drop
}

type Default struct {
	ID   goe.Type[string] `goe:"default:'Default'"`
	Name goe.Type[string]
	goe.EntityMap[Default, DefaultModel]
}

type DefaultModel struct {
	ID   string
	Name string
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
	Insert         *Insert
	Page           *Page
	Default        *Default
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
	db, err := goe.Open[Database](postgres.Open("user=postgres password=postgres host=localhost port=5432 database=postgres", postgres.NewConfig(postgres.Config{
		//Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	})))
	if err != nil {
		return nil, err
	}
	err = db.Migrate().AutoMigrate()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func SetupSqlite() (*Database, error) {
	var err error
	db, err := goe.Open[Database](sqlite.Open(filepath.Join(os.TempDir(), "goe.db"), sqlite.NewConfig(
		sqlite.Config{
			Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
			ConnectionHook: func(conn sqlite.ExecQuerierContext, dsn string) error {
				conn.ExecContext(context.Background(), "PRAGMA foreign_keys = OFF;", nil)
				return nil
			},
		},
	)))
	if err != nil {
		return nil, err
	}
	err = db.Migrate().AutoMigrate()
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
			raceDb.Close()
		}()
	}
	wg.Wait()
}

func TestMigrate(t *testing.T) {
	db, err := Setup()
	if err != nil {
		t.Fatalf("Expected a connection, got error %v", err)
	}
	err = db.Migrate().OnTable("Insert").RenameColumn("Name", "NewName")
	if err != nil {
		t.Fatalf("Expected rename column, got error %v", err)
	}

	err = db.Migrate().OnTable("Insert").DropColumn("NewName")
	if err != nil {
		t.Fatalf("Expected drop column, got error %v", err)
	}

	err = db.Migrate().OnTable("Insert").RenameTable("NewInsert")
	if err != nil {
		t.Fatalf("Expected rename table, got error %v", err)
	}

	err = db.Migrate().OnTable("NewInsert").DropTable()
	if err != nil {
		t.Fatalf("Expected drop table NewInsert, got error %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = db.Migrate().AutoMigrateContext(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected context.Canceled, got %v", err)
	}

	err = db.Migrate().OnSchema("DropSchema").OnTable("Drop").RenameColumn("Name", "NewName")
	if err != nil {
		t.Fatalf("Expected rename column, got error %v", err)
	}

	err = db.Migrate().OnSchema("DropSchema").OnTable("Drop").DropColumn("NewName")
	if err != nil {
		t.Fatalf("Expected drop column, got error %v", err)
	}

	err = db.Migrate().OnSchema("DropSchema").OnTable("Drop").DropTable()
	if err != nil {
		t.Fatalf("Expected drop table DropSchema.Drop, got error %v", err)
	}
}
