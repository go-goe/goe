package tests_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olauro/goe"
	"github.com/olauro/postgres"
)

type Animal struct {
	Id          int
	Name        string
	IdHabitat   *uuid.UUID `goe:"table:Habitat"`
	IdInfo      *[]byte    `goe:"table:Info"`
	Foods       []Food
	AnimalFoods []AnimalFood
}

type AnimalFood struct {
	IdAnimal int       `goe:"pk"`
	IdFood   uuid.UUID `goe:"pk"`
}

type Food struct {
	Id          uuid.UUID
	Name        string
	Animals     []Animal
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
	Name       string
	NameStatus string
	IdStatus   int
}

type Status struct {
	Id   int
	Name string
}

type User struct {
	Id        int
	Name      string
	Email     string `goe:"index(unique n:idx_email)"`
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
	*goe.DB
}

var db *Database

func SetupPostgres() (*Database, error) {
	if db != nil {
		return db, nil
	}
	db = &Database{DB: &goe.DB{}}
	err := goe.Open(db, postgres.Open("user=postgres password=postgres host=localhost port=5432 database=postgres"), goe.Config{})
	if err != nil {
		return nil, err
	}
	err = db.Migrate(goe.MigrateFrom(db))
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
