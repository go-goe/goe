package tests_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-goe/goe"
	"github.com/go-goe/postgres"
	"github.com/google/uuid"
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
	Id      uuid.UUID
	Name    string
	Float32 float32
	Float64 float64
	Today   time.Time
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint    uint
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Bool    bool
	Byte    []byte
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

type Database struct {
	Animal         *Animal
	AnimalFood     *AnimalFood
	Food           *Food
	Habitat        *Habitat
	Info           *Info
	Status         *Status
	Weather        *Weather
	User           *User
	UserRole       *UserRole
	Role           *Role
	Flag           *Flag
	Person         *Person
	PersonJobTitle *PersonJobTitle
	JobTitle       *JobTitle
	Exam           *Exam
	Select         *Select
	Page           *Page
	*goe.DB
}

var db *Database

func SetupPostgres() (*Database, error) {
	if db != nil {
		return db, nil
	}
	var err error
	db, err = goe.Open[Database](postgres.Open("user=postgres password=postgres host=localhost port=5432 database=postgres", postgres.Config{}))
	if err != nil {
		return nil, err
	}
	err = goe.AutoMigrate(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestConnection(t *testing.T) {
	_, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected Postgres Connection, got error %v", err)
	}
}

func TestTx(t *testing.T) {
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

func TestMigrate(t *testing.T) {
	_, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected Postgres Connection, got error %v", err)
	}

	err = goe.RenameColumn(db, "Select", "Name", "NewName")
	if err != nil {
		t.Fatalf("Expected rename column, got error %v", err)
	}

	err = goe.DropColumn(db, "Select", "NewName")
	if err != nil {
		t.Fatalf("Expected drop column, got error %v", err)
	}

	err = goe.DropTable(db, "Select")
	if err != nil {
		t.Fatalf("Expected drop table Select, got error %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = goe.AutoMigrateContext(ctx, db)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected context.Canceled, got %v", err)
	}
}

func TestRace(t *testing.T) {
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			raceDb, _ := goe.Open[Database](postgres.Open("user=postgres password=postgres host=localhost port=5432 database=postgres", postgres.Config{}))
			goe.Close(raceDb)
		}()
	}
	wg.Wait()
}
