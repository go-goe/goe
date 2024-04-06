package main

import (
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lauro-ss/goe"
)

type Produto struct {
	Id         string `goe:"pk;t:uuid"`
	Name       string `goe:"t:varchar(20)"`
	Categorias []Categoria
}

type Categoria struct {
	Id            string `goe:"pk;t:uuid"`
	Name          string `goe:"t:varchar(20)"`
	Produtos      []Produto
	Subcategorias []Subcategoria
}

type Subcategoria struct {
	Id         string `goe:"pk;t:uuid"`
	Name       string `goe:"t:varchar(20)"`
	Categorias []Categoria
}

type Animal struct {
	IdAnimal string `goe:"pk;t:uuid"`
	Emoji    string
	Name     string
	Foods    []Food
	Status   []Status
}

type Status struct {
	Id   string
	Name string
	Animal
}

type Food struct {
	IdFood  string `goe:"pk;t:uuid"`
	Name    string
	Animals []Animal
	Emoji   string
}

type AnimalDb struct {
	IdAnimal goe.Pk
	Name     goe.Att
	Emoji    goe.Att
}

type StatusDb struct {
	Id   goe.Pk
	Name goe.Att
}

type AnimalFood struct {
	IdAnimal goe.Pk
	IdFood   goe.Pk
}

type FoodDb struct {
	IdFood goe.Pk
	Name   goe.Att
	Emoji  goe.Att
}

// TODO: Check if field exists
type Database struct {
	Animal     AnimalDb
	Food       FoodDb
	Status     StatusDb
	AnimalFood AnimalFood
	*goe.DB
}

func main() {

	// db := goe.Connect("database_conection", goe.Config{MigrationsPath: "./Migrations"})
	// db.Migrate(&Produto{})
	// db.Migrate(&Categoria{})
	// db.Migrate(&Subcategoria{})
	// db := &Database{
	// 	Animal: AnimalDb{
	// 		Id: goe.MapAttribute(&Animal{}, "Id"),
	// 	},
	// }

	db := &Database{DB: &goe.DB{}}
	//goe.Map(db.Animal, &Animal{})
	//goe.Connect(db)
	goe.Map(db, Status{})
	goe.Map(db, Animal{})
	goe.Map(db, Food{})

	fmt.Println(db.Animal.IdAnimal)
	fmt.Printf("%p \n", db.Animal.IdAnimal)
	fmt.Println(db.Status.Id)
	fmt.Println(db.Status.Name)
	fmt.Printf("%p \n", db.Status.Id)
	// err := goe.Map(&db.Animal, Animal{})
	// fmt.Println(err)
	//fmt.Printf("%p \n", db.Animal.IdAnimal.Fk["Food"])
	// fmt.Printf("%p Food \n", db.Food.IdFood)
	// // fmt.Println(db.Animal.Name)
	// fmt.Println("Next")
	// fmt.Println(db.AnimalFood.IdAnimal)
	// fmt.Println(db.Animal.IdAnimal)
	// fmt.Printf("%p, %p \n", db.Animal.IdAnimal, db.AnimalFood.IdAnimal)
	// fmt.Printf("%p \n", db.AnimalFood.IdFood)
	// fmt.Println(db.AnimalFood.IdFood)
	// fmt.Println(db.Food.IdFood)
	// fmt.Printf("%p \n", db.Animal.IdAnimal)
	// fmt.Println(db.AnimalFood.IdFood)
	// fmt.Println(db.AnimalFood.IdAnimal)
	// fmt.Printf("%p \n", db.Food.IdFood)
	// fmt.Println(db.Food.IdFood)
	// fmt.Println(db.Animal.IdAnimal)
	// fmt.Printf("%p \n", db.AnimalFood.IdAnimal)
	// fmt.Printf("%p \n", db.AnimalFood.IdFood)
	// fmt.Printf("%p Animal \n", db.Animal.IdAnimal)
	CheckManyToMany(db)
	// fmt.Println(db.Animal.Emoji, db.Food.Emoji)
	//db.Select(db.Animal.IdAnimal)
	db.Select(db.Food.IdFood)
	db.Open("pgx", "user=app password=123456 host=localhost port=5432 database=appanimal sslmode=disable")

	// ids := make([]string, 10)

	//works
	// var ids []string
	// db.Select(db.Animal.Id).Result(&ids)
	// fmt.Println(db.Errors())
	// fmt.Println(ids)

	// var animals []Animal
	// db.Result(&animals)
	// fmt.Println(db.Erros)
	// fmt.Println(animals)

	//db.Select(db.Animal.Id).Where(db.Animal.Id.Equals("1"))

	// db.SetTable(&Produto{})
	// db.SetTable(&Categoria{})
	// "db.Get(&users).Join('Categoria')"
	// "db.Select(&user)"
	// "db.Select('Id','Name', '')"
}

func CheckManyToMany(db *Database) {
	ap := fmt.Sprintf("%p", db.Animal.IdAnimal)
	fp := fmt.Sprintf("%p", db.Food.IdFood)
	f := fmt.Sprint(db.AnimalFood.IdFood)
	a := fmt.Sprint(db.AnimalFood.IdAnimal)
	if !strings.Contains(f, ap) && !strings.Contains(a, fp) {
		fmt.Println("Fail on " + ap + " " + f)
		fmt.Println("Fail on " + fp + " " + a)
	}

	ap = fmt.Sprintf("%p", db.AnimalFood.IdAnimal)
	fp = fmt.Sprintf("%p", db.AnimalFood.IdFood)
	f = fmt.Sprint(db.Food.IdFood)
	a = fmt.Sprint(db.Animal.IdAnimal)
	if !strings.Contains(f, fp) && !strings.Contains(a, ap) {
		fmt.Println("Fail pointer " + ap + " on " + a)
		fmt.Println("Fail pointer " + fp + " on " + f)
	}
}
