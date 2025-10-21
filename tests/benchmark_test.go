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
var size int = 100

func BenchmarkSelect(b *testing.B) {
	db, _ := Setup()

	goe.Delete(db.AnimalFood).All()
	goe.Delete(db.Animal).All()

	animals = make([]Animal, size)
	goe.Insert(db.Animal).All(animals)

	for b.Loop() {
		animals = make([]Animal, 0)
		for row := range goe.List(db.Animal).Rows() {
			animals = append(animals, row)
		}
	}
}

func BenchmarkSelectRaw(b *testing.B) {
	db, _ := Setup()

	goe.Delete(db.AnimalFood).All()
	goe.Delete(db.Animal).All()

	animals = make([]Animal, size)
	goe.Insert(db.Animal).All(animals)

	for b.Loop() {
		rows, _ := db.DB.RawQueryContext(context.Background(), "select a.id, a.name, a.id_info, a.id_habitat from animals a;")
		defer rows.Close()

		var a Animal
		animals = make([]Animal, 0)
		for rows.Next() {
			rows.Scan(&a.Id, &a.Name, &a.InfoId, &a.HabitatId)
			animals = append(animals, a)
		}
	}
}

var foods []Food

func BenchmarkJoin(b *testing.B) {
	db, _ := Setup()

	goe.Delete(db.Weather)
	goe.Delete(db.Habitat)
	goe.Delete(db.AnimalFood).All()
	goe.Delete(db.Animal).All()
	goe.Delete(db.Food).All()

	w := Weather{Name: "Weather"}
	goe.Insert(db.Weather).One(&w)

	h := Habitat{Id: uuid.New(), Name: "Habitat", WeatherId: w.Id}
	goe.Insert(db.Habitat).One(&h)

	a := Animal{Name: "Animal", HabitatId: &h.Id}
	goe.Insert(db.Animal).One(&a)

	f := Food{Id: uuid.New(), Name: "Food"}
	goe.Insert(db.Food).One(&f)

	af := AnimalFood{AnimalId: a.Id, FoodId: f.Id}
	goe.Insert(db.AnimalFood).One(&af)

	for b.Loop() {
		foods = make([]Food, 0)

		for row := range goe.List(db.Food).
			Joins(
				join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
				join.Join[int](&db.AnimalFood.AnimalId, &db.Animal.Id),
				join.Join[uuid.UUID](&db.Animal.HabitatId, &db.Habitat.Id),
				join.Join[int](&db.Habitat.WeatherId, &db.Weather.Id),
			).
			Where(
				where.And(where.Equals(&db.Food.Id, f.Id), where.Equals(&db.Food.Name, f.Name)),
			).
			Rows() {
			foods = append(foods, row)
		}
	}
}

func BenchmarkJoinSql(b *testing.B) {
	db, _ := Setup()

	goe.Delete(db.Weather)
	goe.Delete(db.Habitat)
	goe.Delete(db.AnimalFood).All()
	goe.Delete(db.Animal).All()
	goe.Delete(db.Food).All()

	w := Weather{Name: "Weather"}
	goe.Insert(db.Weather).One(&w)

	h := Habitat{Id: uuid.New(), Name: "Habitat", WeatherId: w.Id}
	goe.Insert(db.Habitat).One(&h)

	a := Animal{Name: "Animal", HabitatId: &h.Id}
	goe.Insert(db.Animal).One(&a)

	f := Food{Id: uuid.New(), Name: "Food"}
	goe.Insert(db.Food).One(&f)

	af := AnimalFood{AnimalId: a.Id, FoodId: f.Id}
	goe.Insert(db.AnimalFood).One(&af)

	for b.Loop() {

		rows, _ := db.DB.RawQueryContext(context.Background(), `select f.id, f.name from foods f
						join animal_foods af on f.id = af.id_food
						join animals a on af.id_animal = a.id
						join habitats h on a.id_habitat = h.id
						join weathers w on h.id_weather = w.id
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
