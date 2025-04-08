package tests_test

import (
	"context"
	"testing"

	"github.com/go-goe/goe"
	"github.com/go-goe/goe/query/join"
	"github.com/go-goe/goe/query/where"
	"github.com/google/uuid"
)

var animals []Animal
var size int = 10000

func BenchmarkSelect(b *testing.B) {
	db, _ := SetupPostgres()

	goe.Delete(db.AnimalFood).Wheres()
	goe.Delete(db.Animal).Wheres()

	animals = make([]Animal, size)
	goe.Insert(db.Animal).All(animals)

	for b.Loop() {
		animals = make([]Animal, 0)
		for row := range goe.Select(db.Animal).From(db.Animal).Rows() {
			animals = append(animals, row)
		}
	}
}

func BenchmarkSelectRaw(b *testing.B) {
	db, _ := SetupPostgres()

	goe.Delete(db.AnimalFood).Wheres()
	goe.Delete(db.Animal).Wheres()

	animals = make([]Animal, size)
	goe.Insert(db.Animal).All(animals)

	for b.Loop() {
		rows, _ := db.DB.RawQueryContext(context.Background(), "select a.id, a.name, a.idinfo, a.idhabitat from animals a;")
		defer rows.Close()

		var a Animal
		animals = make([]Animal, 0)
		for rows.Next() {
			rows.Scan(&a.Id, &a.Name, &a.IdInfo, &a.IdHabitat)
			animals = append(animals, a)
		}
	}
}

var foods []Food

func BenchmarkJoin(b *testing.B) {
	db, _ := SetupPostgres()

	goe.Delete(db.Weather)
	goe.Delete(db.Habitat)
	goe.Delete(db.AnimalFood).Wheres()
	goe.Delete(db.Animal).Wheres()
	goe.Delete(db.Food).Wheres()

	w := Weather{Name: "Weather"}
	goe.Insert(db.Weather).One(&w)

	h := Habitat{Id: uuid.New(), Name: "Habitat", IdWeather: w.Id}
	goe.Insert(db.Habitat).One(&h)

	a := Animal{Name: "Animal", IdHabitat: &h.Id}
	goe.Insert(db.Animal).One(&a)

	f := Food{Id: uuid.New(), Name: "Food"}
	goe.Insert(db.Food).One(&f)

	af := AnimalFood{IdAnimal: a.Id, IdFood: f.Id}
	goe.Insert(db.AnimalFood).One(&af)

	for b.Loop() {
		foods = make([]Food, 0)

		for row := range goe.Select(db.Food).From(db.Food).
			Joins(
				join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
				join.Join[int](&db.AnimalFood.IdAnimal, &db.Animal.Id),
				join.Join[uuid.UUID](&db.Animal.IdHabitat, &db.Habitat.Id),
				join.Join[int](&db.Habitat.IdWeather, &db.Weather.Id),
			).
			Wheres(
				where.Equals(&db.Food.Id, f.Id),
				where.And(),
				where.Equals(&db.Food.Name, f.Name),
			).
			Rows() {
			foods = append(foods, row)
		}
	}
}

func BenchmarkJoinSql(b *testing.B) {
	db, _ := SetupPostgres()

	goe.Delete(db.Weather)
	goe.Delete(db.Habitat)
	goe.Delete(db.AnimalFood).Wheres()
	goe.Delete(db.Animal).Wheres()
	goe.Delete(db.Food).Wheres()

	w := Weather{Name: "Weather"}
	goe.Insert(db.Weather).One(&w)

	h := Habitat{Id: uuid.New(), Name: "Habitat", IdWeather: w.Id}
	goe.Insert(db.Habitat).One(&h)

	a := Animal{Name: "Animal", IdHabitat: &h.Id}
	goe.Insert(db.Animal).One(&a)

	f := Food{Id: uuid.New(), Name: "Food"}
	goe.Insert(db.Food).One(&f)

	af := AnimalFood{IdAnimal: a.Id, IdFood: f.Id}
	goe.Insert(db.AnimalFood).One(&af)

	for b.Loop() {

		rows, _ := db.DB.RawQueryContext(context.Background(), `select f.id, f.name from foods f
						join animalfoods af on f.id = af.idfood
						join animals a on af.idanimal = a.id
						join habitats h on a.idhabitat = h.id
						join weathers w on h.idweather = w.id
						where f.id = $1 and f.name = $2;`, f.Id, f.Name)
		defer rows.Close()

		foods = make([]Food, 0)
		var food Food
		for rows.Next() {
			rows.Scan(&food.Id, &food.Name)
			foods = append(foods, food)
		}
	}
}
