package tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olauro/goe"
	"github.com/olauro/postgres"
)

type Animal struct {
	Name        string     `goe:"index"`
	IdHabitat   *uuid.UUID `goe:"table:Habitat"`
	IdInfo      *[]byte    `goe:"table:Info"`
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
	IdUser  int
	IdRole  int
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
	IdPerson   int `goe:"pk"`
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
}

var db *Database

func SetupPostgres() (*Database, error) {
	if db != nil {
		return db, nil
	}
	var err error
	db, err = goe.Open[Database](postgres.Open("user=postgres password=postgres host=localhost port=5432 database=postgres"), goe.Config{})
	if err != nil {
		return nil, err
	}
	err = goe.AutoMigrate(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestPostgresConnection(t *testing.T) {
	_, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected Postgres Connection, got error %v", err)
	}
}

func TestPostgresMigrate(t *testing.T) {
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
