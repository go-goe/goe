package tests_test

import (
	"context"
	"errors"
	"iter"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olauro/goe"
	"github.com/olauro/goe/query"
)

func TestPostgresSelect(t *testing.T) {
	db, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}
	goeDb, err := goe.GetGoeDatabase(db)
	if err != nil {
		t.Fatalf("Expected goe database, got error: %v", err)
	}

	err = goe.Delete(db.AnimalFood).Where()
	if err != nil {
		t.Fatalf("Expected delete AnimalFood, got error: %v", err)
	}
	err = goe.Delete(db.Flag).Where()
	if err != nil {
		t.Fatalf("Expected delete flags, got error: %v", err)
	}
	err = goe.Delete(db.Animal).Where()
	if err != nil {
		t.Fatalf("Expected delete animals, got error: %v", err)
	}
	err = goe.Delete(db.Food).Where()
	if err != nil {
		t.Fatalf("Expected delete foods, got error: %v", err)
	}
	err = goe.Delete(db.Habitat).Where()
	if err != nil {
		t.Fatalf("Expected delete habitats, got error: %v", err)
	}
	err = goe.Delete(db.Info).Where()
	if err != nil {
		t.Fatalf("Expected delete infos, got error: %v", err)
	}
	err = goe.Delete(db.Status).Where()
	if err != nil {
		t.Fatalf("Expected delete status, got error: %v", err)
	}
	err = goe.Delete(db.UserRole).Where()
	if err != nil {
		t.Fatalf("Expected delete user roles, got error: %v", err)
	}
	err = goe.Delete(db.User).Where()
	if err != nil {
		t.Fatalf("Expected delete users, got error: %v", err)
	}
	err = goe.Delete(db.Role).Where()
	if err != nil {
		t.Fatalf("Expected delete roles, got error: %v", err)
	}
	err = goe.Delete(db.PersonJobTitle).Where()
	if err != nil {
		t.Fatalf("Expected delete personJobs, got error: %v", err)
	}
	err = goe.Delete(db.JobTitle).Where()
	if err != nil {
		t.Fatalf("Expected delete jobs, got error: %v", err)
	}
	err = goe.Delete(db.Person).Where()
	if err != nil {
		t.Fatalf("Expected delete persons, got error: %v", err)
	}
	err = goe.Delete(db.Exam).Where()
	if err != nil {
		t.Fatalf("Expected delete exams, got error: %v", err)
	}

	weathers := []Weather{
		{Name: "Hot"},
		{Name: "Cold"},
		{Name: "Wind"},
		{Name: "Nice"},
		{Name: "Ocean"},
	}
	err = goe.Insert(db.Weather).All(weathers)
	if err != nil {
		t.Fatalf("Expected insert weathers, got error: %v", err)
	}

	habitats := []Habitat{
		{Id: uuid.New(), Name: "City", IdWeather: weathers[0].Id, NameWeather: "Test"},
		{Id: uuid.New(), Name: "Jungle", IdWeather: weathers[3].Id},
		{Id: uuid.New(), Name: "Savannah", IdWeather: weathers[0].Id},
		{Id: uuid.New(), Name: "Ocean", IdWeather: weathers[2].Id},
	}
	err = goe.Insert(db.Habitat).All(habitats)
	if err != nil {
		t.Fatalf("Expected insert habitats, got error: %v", err)
	}

	status := []Status{
		{Name: "Cat Alive"},
		{Name: "Dog Alive"},
		{Name: "Big Dog Alive"},
	}

	err = goe.Insert(db.Status).All(status)
	if err != nil {
		t.Fatalf("Expected insert habitats, got error: %v", err)
	}

	infos := []Info{
		{Id: uuid.New().NodeID(), Name: "Little Cat", IdStatus: status[0].Id, NameStatus: "Test"},
		{Id: uuid.New().NodeID(), Name: "Big Dog", IdStatus: status[2].Id},
	}
	err = goe.Insert(db.Info).All(infos)
	if err != nil {
		t.Fatalf("Expected insert infos, got error: %v", err)
	}

	animals := []Animal{
		{Name: "Cat", IdHabitat: &habitats[0].Id, IdInfo: &infos[0].Id},
		{Name: "Dog", IdHabitat: &habitats[0].Id, IdInfo: &infos[1].Id},
		{Name: "Forest Cat", IdHabitat: &habitats[1].Id},
		{Name: "Little cat", IdHabitat: &habitats[1].Id},
		{Name: "Bear", IdHabitat: &habitats[1].Id},
		{Name: "Lion", IdHabitat: &habitats[2].Id},
		{Name: "Puma", IdHabitat: &habitats[1].Id},
		{Name: "Snake", IdHabitat: &habitats[1].Id},
		{Name: "Whale"},
	}
	err = goe.Insert(db.Animal).All(animals)
	if err != nil {
		t.Fatalf("Expected insert animals, got error: %v", err)
	}

	foods := []Food{{Id: uuid.New(), Name: "Meat"}, {Id: uuid.New(), Name: "Grass"}}
	err = goe.Insert(db.Food).All(foods)
	if err != nil {
		t.Fatalf("Expected insert foods, got error: %v", err)
	}

	animalFoods := []AnimalFood{
		{IdFood: foods[0].Id, IdAnimal: animals[0].Id},
		{IdFood: foods[0].Id, IdAnimal: animals[1].Id}}
	err = goe.Insert(db.AnimalFood).All(animalFoods)
	if err != nil {
		t.Fatalf("Expected insert animalFoods, got error: %v", err)
	}

	users := []User{
		{Name: "Lauro Santana", Email: "lauro@email.com"},
		{Name: "John Constantine", Email: "hunter@email.com"},
		{Name: "Harry Potter", Email: "harry@email.com"},
	}
	err = goe.Insert(db.User).All(users)
	if err != nil {
		t.Fatalf("Expected insert users, got error: %v", err)
	}

	roles := []Role{
		{Name: "Administrator"},
		{Name: "User"},
		{Name: "Mid-Level"},
	}
	err = goe.Insert(db.Role).All(roles)
	if err != nil {
		t.Fatalf("Expected insert roles, got error: %v", err)
	}

	tt := time.Now().AddDate(0, 0, 10)
	userRoles := []UserRole{
		{IdUser: users[0].Id, IdRole: roles[0].Id, EndDate: &tt},
		{IdUser: users[1].Id, IdRole: roles[2].Id},
	}
	err = goe.Insert(db.UserRole).All(userRoles)
	if err != nil {
		t.Fatalf("Expected insert user roles, got error: %v", err)
	}

	persons := []Person{
		{Name: "Jhon"},
		{Name: "Laura"},
		{Name: "Luana"},
	}
	err = goe.Insert(db.Person).All(persons)
	if err != nil {
		t.Fatalf("Expected insert persons, got error: %v", err)
	}

	jobs := []JobTitle{
		{Name: "Developer"},
		{Name: "Designer"},
	}
	err = goe.Insert(db.JobTitle).All(jobs)
	if err != nil {
		t.Fatalf("Expected insert jobs, got error: %v", err)
	}

	personJobs := []PersonJobTitle{
		{IdPerson: persons[0].Id, IdJobTitle: jobs[0].Id, CreatedAt: time.Now()},
		{IdPerson: persons[1].Id, IdJobTitle: jobs[0].Id, CreatedAt: time.Now()},
		{IdPerson: persons[2].Id, IdJobTitle: jobs[1].Id, CreatedAt: time.Now()},
	}
	err = goe.Insert(db.PersonJobTitle).All(personJobs)
	if err != nil {
		t.Fatalf("Expected insert personJobs, got error: %v", err)
	}

	exams := []Exam{
		{Score: 9.9, Minimum: 5.5},
		{Score: 4.9, Minimum: 5.5},
		{Score: 5.5, Minimum: 5.5},
	}
	err = goe.Insert(db.Exam).All(exams)
	if err != nil {
		t.Fatalf("Expected insert exams, got error: %v", err)
	}

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Select",
			testCase: func(t *testing.T) {
				a := runSelect(t, goe.Select(db.Animal).From(db.Animal).Rows())
				if len(a) != len(animals) {
					t.Errorf("Expected %v animals, got %v", len(animals), len(a))
				}
			},
		},
		{
			desc: "Find",
			testCase: func(t *testing.T) {
				var a *Animal
				a, err = goe.Find(db.Animal, Animal{Id: animals[0].Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}
				if a.Name != animals[0].Name {
					t.Errorf("Expected a %v, got %v", animals[0].Name, a.Name)
				}
			},
		},
		{
			desc: "Find_Composed_Pk",
			testCase: func(t *testing.T) {
				var a *AnimalFood
				a, err = goe.Find(db.AnimalFood, AnimalFood{IdAnimal: animals[0].Id, IdFood: foods[0].Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}
				if a.IdAnimal != animals[0].Id {
					t.Errorf("Expected a %v, got %v", animals[0].Id, a.IdAnimal)
				}
				if a.IdFood != foods[0].Id {
					t.Errorf("Expected a %v, got %v", foods[0].Id, a.IdFood)
				}
			},
		},
		{
			desc: "Select_Where_Greater",
			testCase: func(t *testing.T) {
				e := runSelect(t, goe.Select(db.Exam).From(db.Exam).
					Where(query.GreaterArg(&db.Exam.Score, &db.Exam.Minimum)).Rows())
				if len(e) != 1 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}

				e = nil
				e = runSelect(t, goe.Select(db.Exam).From(db.Exam).
					Where(query.Greater(&db.Exam.Score, float32(5.5))).Rows())
				if len(e) != 1 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}
			},
		},
		{
			desc: "Select_Where_GreaterEquals",
			testCase: func(t *testing.T) {
				e := runSelect(t, goe.Select(db.Exam).From(db.Exam).
					Where(query.GreaterEqualsArg(&db.Exam.Score, &db.Exam.Minimum)).Rows())
				if len(e) != 2 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}

				e = nil
				e = runSelect(t, goe.Select(db.Exam).From(db.Exam).
					Where(query.GreaterEquals(&db.Exam.Score, float32(5.5))).Rows())
				if len(e) != 2 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}
			},
		},
		{
			desc: "Select_Where_Less",
			testCase: func(t *testing.T) {
				e := runSelect(t, goe.Select(db.Exam).From(db.Exam).
					Where(query.LessArg(&db.Exam.Score, &db.Exam.Minimum)).Rows())
				if len(e) != 1 {
					t.Errorf("Expected %v, got %v", 1, len(e))
				}

				e = nil
				e = runSelect(t, goe.Select(db.Exam).From(db.Exam).
					Where(query.Less(&db.Exam.Score, float32(5.5))).Rows())
				if len(e) != 1 {
					t.Errorf("Expected %v, got %v", 1, len(e))
				}
			},
		},
		{
			desc: "Select_Where_LessEquals",
			testCase: func(t *testing.T) {
				e := runSelect(t, goe.Select(db.Exam).From(db.Exam).
					Where(query.LessEqualsArg(&db.Exam.Score, &db.Exam.Minimum)).Rows())
				if len(e) != 2 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}

				e = nil
				e = runSelect(t, goe.Select(db.Exam).From(db.Exam).
					Where(query.LessEquals(&db.Exam.Score, float32(5.5))).Rows())
				if len(e) != 2 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}
			},
		},
		{
			desc: "Select_Where_Like",
			testCase: func(t *testing.T) {
				a := runSelect(t, goe.Select(db.Animal).
					From(db.Animal).Where(query.Like(&db.Animal.Name, "%Cat%")).Rows())
				if len(a) != 2 {
					t.Errorf("Expected %v animals, got %v", 2, len(a))
				}
			},
		},
		{
			desc: "Select_Where_Custom_Operation",
			testCase: func(t *testing.T) {
				if goeDb.Driver.Name() == "PostgreSQL" {
					qr := goe.Select(db.Animal).From(db.Animal).Where(query.NewOperator(&db.Animal.Name, "ILIKE", "%CAT%")).Rows()
					a := runSelect(t, qr)
					if len(a) != 3 {
						t.Errorf("Expected %v animals, got %v", 3, len(a))
					}
				}
			},
		},
		{
			desc: "Select_Where_Equals_Nil",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Animal).From(db.Animal).Where(query.Equals[*uuid.UUID](&db.Animal.IdHabitat, nil)).Rows()
				a := runSelect(t, qr)
				if len(a) != 1 {
					t.Errorf("Expected %v animals, got %v", 1, len(a))
				}
			},
		},
		{
			desc: "Select_Where_NotEquals_Nil",
			testCase: func(t *testing.T) {
				var bb *[]byte
				qr := goe.Select(db.Animal).From(db.Animal).Where(query.NotEquals(&db.Animal.IdInfo, bb)).Rows()
				a := runSelect(t, qr)
				if len(a) != len(infos) {
					t.Errorf("Expected %v animals, got %v", len(infos), len(a))
				}
			},
		},
		{
			desc: "Find_Not_Found",
			testCase: func(t *testing.T) {
				_, err = goe.Find(db.Animal, Animal{Id: -1})
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}
			},
		},
		{
			desc: "Select_Order_By_Asc",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Animal).From(db.Animal).OrderByAsc(&db.Animal.Id).Rows()
				a := runSelect(t, qr)
				if a[0].Id > a[1].Id {
					t.Errorf("Expected animals order by asc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Order_By_Desc",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Animal).From(db.Animal).OrderByDesc(&db.Animal.Id).Rows()
				a := runSelect(t, qr)
				if a[0].Id < a[1].Id {
					t.Errorf("Expected animals order by desc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Page",
			testCase: func(t *testing.T) {
				var pageSize uint = 5
				qr := goe.Select(db.Animal).From(db.Animal).Page(1, pageSize).Rows()
				a := runSelect(t, qr)
				if len(a) != int(pageSize) {
					t.Errorf("Expected %v animals, got %v", pageSize, len(a))
				}
			},
		},
		{
			desc: "Select_Join",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Animal).
					From(db.Animal).
					Joins(
						query.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
						query.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
					).Rows()
				a := runSelect(t, qr)

				if len(a) != len(animalFoods) {
					t.Errorf("Expected 1 animal, got %v", len(a))
				}
				if a[0].Name != animals[0].Name {
					t.Errorf("Expected %v, got %v", animals[0].Name, a[0].Name)
				}
			},
		},
		{
			desc: "Select_Join_Implicit",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Animal).
					From(db.Animal, db.AnimalFood, db.Food).
					Where(
						query.EqualsArg(&db.Animal.Id, &db.AnimalFood.IdAnimal),
						query.And(),
						query.EqualsArg(&db.Food.Id, &db.AnimalFood.IdFood)).Rows()
				a := runSelect(t, qr)

				if len(a) != len(animalFoods) {
					t.Errorf("Expected 1 animal, got %v", len(a))
				}
				if a[0].Name != animals[0].Name {
					t.Errorf("Expected %v, got %v", animals[0].Name, a[0].Name)
				}
			},
		},
		{
			desc: "Select_Join_Where",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Food).
					From(db.Food).
					Joins(
						query.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
						query.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
					).
					Where(
						query.Equals(&db.Animal.Name, animals[0].Name)).Rows()
				f := runSelect(t, qr)

				if len(f) != 1 {
					t.Fatalf("Expected 1 food, got %v", len(f))
				}
				if f[0].Name != foods[0].Name {
					t.Errorf("Expected %v, got %v", foods[0].Name, f[0].Name)
				}
			},
		},
		{
			desc: "Select_Join_Order_By_Asc",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Animal).
					From(db.Animal).
					Joins(
						query.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
						query.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
					).
					OrderByAsc(&db.Animal.Id).Rows()
				a := runSelect(t, qr)
				if a[0].Id > a[1].Id {
					t.Errorf("Expected animals order by asc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Join_Order_By_Desc",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Animal).
					From(db.Animal).
					Joins(
						query.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
						query.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
					).
					OrderByDesc(&db.Animal.Id).Rows()
				a := runSelect(t, qr)
				if a[0].Id < a[1].Id {
					t.Errorf("Expected animals order by desc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Join_Where_Order_By_Asc",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Animal).
					From(db.Animal).
					Joins(
						query.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
						query.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
					).
					Where(
						query.Equals(&db.Food.Id, foods[0].Id),
					).OrderByAsc(&db.Animal.Id).Rows()
				a := runSelect(t, qr)

				if len(a) != 2 {
					t.Fatalf("Expected 2 animals, got %v", len(a))
				}
				if a[0].Id > a[1].Id {
					t.Errorf("Expected animals order by asc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Join_Where_Order_By_Desc",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Animal).
					From(db.Animal).
					Joins(
						query.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
						query.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
					).
					Where(
						query.Equals(&db.Food.Id, foods[0].Id),
					).OrderByDesc(&db.Animal.Id).Rows()
				a := runSelect(t, qr)

				if len(a) != 2 {
					t.Fatalf("Expected 2 animals, got %v", len(a))
				}
				if a[0].Id < a[1].Id {
					t.Errorf("Expected animals order by desc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Join_Many_To_Many_And_Many_To_One",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Food).
					From(db.Food).
					Joins(
						query.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
						query.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
						query.Join[uuid.UUID](&db.Animal.IdHabitat, &db.Habitat.Id),
					).
					Where(query.Equals(&db.Habitat.Id, habitats[0].Id)).Rows()
				f := runSelect(t, qr)

				if len(f) != 2 {
					t.Errorf("Expected 2, got : %v", len(f))
				}
			},
		},
		{
			desc: "Select_Join_One_To_One",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Animal).From(db.Animal).
					Joins(
						query.Join[[]byte](&db.Animal.IdInfo, &db.Info.Id),
					).Rows()
				a := runSelect(t, qr)

				if len(a) != 2 {
					t.Errorf("Expected 2, got : %v", len(a))
				}
			},
		},
		{
			desc: "Select_Info_Join_Status_One_To_One_And_Many_To_Many",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Info).
					From(db.Info).
					Joins(
						query.Join[int](&db.Status.Id, &db.Info.IdStatus),
						query.Join[[]byte](&db.Animal.IdInfo, &db.Info.Id),
						query.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
						query.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
					).
					Where(query.Equals(&db.Food.Id, foods[0].Id)).Rows()
				s := runSelect(t, qr)

				if len(s) != 2 {
					t.Errorf("Expected 2, got : %v", len(s))
				}
			},
		},
		{
			desc: "Select_Join_Page",
			testCase: func(t *testing.T) {
				var pageSize uint = 2

				qr := goe.Select(db.Animal).From(db.Animal).
					Joins(
						query.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
						query.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
					).
					Page(1, pageSize).Rows()
				a := runSelect(t, qr)

				if len(a) != int(pageSize) {
					t.Errorf("Expected %v animals, got %v", pageSize, len(a))
				}
			},
		},
		{
			desc: "Select_Join_Name",
			testCase: func(t *testing.T) {
				qr := goe.Select(db.Habitat).
					From(db.Habitat).
					Joins(
						query.Join[string](&db.Habitat.Name, &db.Weather.Name),
					).Rows()
				h := runSelect(t, qr)

				if h[0].Name != "Ocean" {
					t.Errorf("Expected Ocean, got : %v", h[0].Name)
				}
			},
		},
		{
			desc: "Select_User_And_Roles",
			testCase: func(t *testing.T) {
				var q []struct {
					User    string
					Role    string
					EndTime *time.Time
				}

				for row, err := range goe.Select(&struct {
					User    *string
					Role    *string
					EndTime **time.Time
				}{
					User:    &db.User.Name,
					Role:    &db.Role.Name,
					EndTime: &db.UserRole.EndDate,
				}).
					From(db.User).
					Joins(
						query.LeftJoin[int](&db.User.Id, &db.UserRole.IdUser),
						query.LeftJoin[int](&db.UserRole.IdRole, &db.Role.Id),
					).
					OrderByAsc(&db.User.Id).Rows() {

					if err != nil {
						t.Fatal(err)
					}

					q = append(q, struct {
						User    string
						Role    string
						EndTime *time.Time
					}{
						User:    goe.SafeGet(row.User),
						Role:    goe.SafeGet(row.Role),
						EndTime: goe.SafeGet(row.EndTime),
					})
				}

				if len(q) != len(users) {
					t.Errorf("Expected %v, got : %v", len(users), len(q))
				}
				if q[0].EndTime == nil {
					t.Errorf("Expected a value, got : %v", q[0].EndTime)
				}
			},
		},
		{
			desc: "Select_User_And_Roles_RightJoin",
			testCase: func(t *testing.T) {
				var q []struct {
					User    string
					Role    string
					EndTime *time.Time
				}

				for row, err := range goe.Select(&struct {
					User    *string
					Role    *string
					EndTime **time.Time
				}{
					User:    &db.User.Name,
					Role:    &db.Role.Name,
					EndTime: &db.UserRole.EndDate,
				}).
					From(db.User).
					Joins(
						query.RightJoin[int](&db.UserRole.IdUser, &db.User.Id),
						query.RightJoin[int](&db.Role.Id, &db.UserRole.IdRole),
					).
					OrderByAsc(&db.User.Id).Rows() {

					if err != nil {
						t.Fatal(err)
					}

					q = append(q, struct {
						User    string
						Role    string
						EndTime *time.Time
					}{
						User:    goe.SafeGet(row.User),
						Role:    goe.SafeGet(row.Role),
						EndTime: goe.SafeGet(row.EndTime),
					})
				}

				if len(q) != len(roles) {
					t.Errorf("Expected %v, got : %v", len(roles), len(q))
				}
				if q[0].EndTime == nil {
					t.Errorf("Expected a value, got : %v", q[0].EndTime)
				}
			},
		},
		{
			desc: "Select_Persons_And_Jobs",
			testCase: func(t *testing.T) {
				pj := []struct {
					JobTitle string
					Person   string
				}{}

				for row, err := range goe.Select(&struct {
					JobTitle *string
					Person   *string
				}{
					JobTitle: &db.JobTitle.Name,
					Person:   &db.Person.Name,
				}).From(db.Person).
					Joins(
						query.Join[int](&db.Person.Id, &db.PersonJobTitle.IdPerson),
						query.Join[int](&db.PersonJobTitle.IdJobTitle, &db.JobTitle.Id),
					).Rows() {

					if err != nil {
						t.Fatal(err)
					}

					pj = append(pj, struct {
						JobTitle string
						Person   string
					}{
						JobTitle: goe.SafeGet(row.JobTitle),
						Person:   goe.SafeGet(row.Person),
					})
				}

				if len(pj) != len(personJobs) {
					t.Errorf("Expected %v, got : %v", len(personJobs), len(pj))
				}
			},
		},
		{
			desc: "Select_Invalid_OrderBy",
			testCase: func(t *testing.T) {
				for _, err := range goe.Select(db.Animal).
					From(db.Animal).OrderByAsc(db.Animal.IdHabitat).Rows() {
					if !errors.Is(err, goe.ErrInvalidOrderBy) {
						t.Errorf("Expected goe.ErrInvalidOrderBy, got error: %v", err)
					}
				}
			},
		},
		{
			desc: "Select_Invalid_Where",
			testCase: func(t *testing.T) {
				for _, err := range goe.Select(db.Animal).
					From(db.Animal).Where(query.Equals(db.Animal.IdHabitat, uuid.New())).Rows() {
					if !errors.Is(err, goe.ErrInvalidWhere) {
						t.Errorf("Expected goe.ErrInvalidWhere, got error: %v", err)
					}
				}
			},
		},
		{
			desc: "Select_Invalid_Arg",
			testCase: func(t *testing.T) {
				goeDb.Config.LogQuery = true
				for _, err := range goe.Select(&struct{}{}).Rows() {
					if !errors.Is(err, goe.ErrInvalidArg) {
						t.Errorf("Expected goe.ErrInvalidArg, got error: %v", err)
					}
				}
				goeDb.Config.LogQuery = false
				for _, err := range goe.Select[any](nil).Rows() {
					if !errors.Is(err, goe.ErrInvalidArg) {
						t.Errorf("Expected goe.ErrInvalidArg, got error: %v", err)
					}
				}
			},
		},
		{
			desc: "Select_Invalid_Table",
			testCase: func(t *testing.T) {
				for _, err := range goe.Select(db.Animal).From(db).Rows() {
					if !errors.Is(err, goe.ErrInvalidArg) {
						t.Errorf("Expected goe.ErrInvalidArg, got error: %v", err)
					}
				}

				for _, err := range goe.Select(db.Animal).From(nil).Rows() {
					if !errors.Is(err, goe.ErrInvalidArg) {
						t.Errorf("Expected goe.ErrInvalidArg, got error: %v", err)
					}
				}
			},
		},
		{
			desc: "Select_Context_Cancel",
			testCase: func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				for _, err := range goe.SelectContext(ctx, db.Animal).From(db.Animal).Rows() {
					if !errors.Is(err, context.Canceled) {
						t.Errorf("Expected a context.Canceled, got error: %v", err)
					}
				}
			},
		},
		{
			desc: "Select_Context_Timeout",
			testCase: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond*1)
				defer cancel()
				for _, err := range goe.SelectContext(ctx, db.Animal).From(db.Animal).Rows() {
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Errorf("Expected a context.DeadlineExceeded, got error: %v", err)
					}
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}

func runSelect[T any](t *testing.T, it iter.Seq2[T, error]) []T {
	rows := make([]T, 0)
	for row, err := range it {
		if err != nil {
			t.Fatalf("Expected a select, got error: %v", err)
		}
		rows = append(rows, row)
	}
	return rows
}
