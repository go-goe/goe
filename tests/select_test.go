package tests_test

import (
	"context"
	"errors"
	"iter"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-goe/goe"
	"github.com/go-goe/goe/query"
	"github.com/go-goe/goe/query/aggregate"
	"github.com/go-goe/goe/query/function"
	"github.com/go-goe/goe/query/join"
	"github.com/go-goe/goe/query/where"
	"github.com/google/uuid"
)

func TestSelect(t *testing.T) {
	db, err := Setup()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}

	if err != nil {
		t.Fatalf("Expected goe database, got error: %v", err)
	}

	err = goe.Delete(db.AnimalFood).All()
	if err != nil {
		t.Fatalf("Expected delete AnimalFood, got error: %v", err)
	}
	err = goe.Delete(db.Flag).All()
	if err != nil {
		t.Fatalf("Expected delete flags, got error: %v", err)
	}
	err = goe.Delete(db.Animal).All()
	if err != nil {
		t.Fatalf("Expected delete animals, got error: %v", err)
	}
	err = goe.Delete(db.Food).All()
	if err != nil {
		t.Fatalf("Expected delete foods, got error: %v", err)
	}
	err = goe.Delete(db.Habitat).All()
	if err != nil {
		t.Fatalf("Expected delete habitats, got error: %v", err)
	}
	err = goe.Delete(db.Info).All()
	if err != nil {
		t.Fatalf("Expected delete infos, got error: %v", err)
	}
	err = goe.Delete(db.Status).All()
	if err != nil {
		t.Fatalf("Expected delete status, got error: %v", err)
	}
	err = goe.Delete(db.UserRole).All()
	if err != nil {
		t.Fatalf("Expected delete user roles, got error: %v", err)
	}
	err = goe.Delete(db.User).All()
	if err != nil {
		t.Fatalf("Expected delete users, got error: %v", err)
	}
	err = goe.Delete(db.Role).All()
	if err != nil {
		t.Fatalf("Expected delete roles, got error: %v", err)
	}
	err = goe.Delete(db.PersonJobTitle).All()
	if err != nil {
		t.Fatalf("Expected delete personJobs, got error: %v", err)
	}
	err = goe.Delete(db.JobTitle).All()
	if err != nil {
		t.Fatalf("Expected delete jobs, got error: %v", err)
	}
	err = goe.Delete(db.Person).All()
	if err != nil {
		t.Fatalf("Expected delete persons, got error: %v", err)
	}
	err = goe.Delete(db.Exam).All()
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
		{Id: uuid.New(), Name: "City", WeatherId: weathers[0].Id, NameWeather: "Test"},
		{Id: uuid.New(), Name: "Jungle", WeatherId: weathers[3].Id},
		{Id: uuid.New(), Name: "Savannah", WeatherId: weathers[0].Id},
		{Id: uuid.New(), Name: "Ocean", WeatherId: weathers[2].Id},
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
		{Id: uuid.New().NodeID(), Name: "Little Cat", StatusId: status[0].Id, NameStatus: "Test"},
		{Id: uuid.New().NodeID(), Name: "Big Dog", StatusId: status[2].Id},
	}
	err = goe.Insert(db.Info).All(infos)
	if err != nil {
		t.Fatalf("Expected insert infos, got error: %v", err)
	}

	animals := []Animal{
		{Name: "Cat", HabitatId: &habitats[0].Id, InfoId: &infos[0].Id},
		{Name: "Dog", HabitatId: &habitats[0].Id, InfoId: &infos[1].Id},
		{Name: "Forest Cat", HabitatId: &habitats[1].Id},
		{Name: "Little cat", HabitatId: &habitats[1].Id},
		{Name: "Bear", HabitatId: &habitats[1].Id},
		{Name: "Lion", HabitatId: &habitats[2].Id},
		{Name: "Puma", HabitatId: &habitats[1].Id},
		{Name: "Snake", HabitatId: &habitats[1].Id},
		{Name: "Whale"},
		{Name: "Wolf"},
		{Name: "Spider"},
		{Name: "Roach"},
		{Name: "Cricket"},
		{Name: "Eagle"},
		{Name: "Falcon"},
		{Name: "Shark"},
		{Name: "Dolphin"},
		{Name: "Turtle"},
		{Name: "Crocodile"},
		{Name: "Frog"},
		{Name: "Elephant"},
		{Name: "Giraffe"},
		{Name: "Kangaroo"},
		{Name: "Koala"},
		{Name: "Penguin"},
		{Name: "Polar Bear"},
		{Name: "Octopus"},
		{Name: "Jellyfish"},
		{Name: "Ant"},
		{Name: "Butterfly"},
		{Name: "Owl"},
		{Name: "Fox"},
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
		{FoodId: foods[0].Id, AnimalId: animals[0].Id},
		{FoodId: foods[0].Id, AnimalId: animals[1].Id}}
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
		{UserId: users[0].Id, RoleId: roles[0].Id, EndDate: &tt},
		{UserId: users[1].Id, RoleId: roles[2].Id},
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
		{PersonId: persons[0].Id, JobTitleId: jobs[0].Id, CreatedAt: time.Now()},
		{PersonId: persons[1].Id, JobTitleId: jobs[0].Id, CreatedAt: time.Now()},
		{PersonId: persons[2].Id, JobTitleId: jobs[1].Id, CreatedAt: time.Now()},
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
				a := runSelect(t, goe.List(db.Animal).Rows())
				if len(a) != len(animals) {
					t.Errorf("Expected %v animals, got %v", len(animals), len(a))
				}
			},
		},
		{
			desc: "Select_Race",
			testCase: func(t *testing.T) {
				var wg sync.WaitGroup
				for range 10 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						goe.List(db.Animal).AsSlice()
					}()
				}
				wg.Wait()
			},
		},
		{
			desc: "List",
			testCase: func(t *testing.T) {
				var a []Animal
				a, err = goe.List(db.Animal).AsSlice()
				if err != nil {
					t.Fatalf("Expected List, got error: %v", err)
				}
				if len(a) != len(animals) {
					t.Errorf("Expected %v animals, got %v", len(animals), len(a))
				}
				a, err = goe.List(db.Animal).
					Joins(
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
						join.Join[uuid.UUID](&db.AnimalFood.FoodId, &db.Food.Id),
					).AsSlice()
				if err != nil {
					t.Fatalf("Expected List, got error: %v", err)
				}
				if len(a) != len(animalFoods) {
					t.Errorf("Expected %v animals, got %v", len(animalFoods), len(a))
				}
			},
		},
		{
			desc: "Select_ToUpper",
			testCase: func(t *testing.T) {
				for row, err := range goe.Select[struct {
					Name      string
					UpperName query.Function[string]
				}](&struct {
					Name      *string
					UpperName *query.Function[string]
				}{
					Name:      &db.Animal.Name,
					UpperName: function.ToUpper(&db.Animal.Name),
				}).Rows() {
					if err != nil {
						t.Fatalf("Expected select, got error: %v", err)
					}
					if strings.ToUpper(row.Name) != row.UpperName.Value {
						t.Fatalf("Expected %v, got: %v", strings.ToUpper(row.Name), row.UpperName.Value)
					}
				}
			},
		},
		{
			desc: "Select_ToLower",
			testCase: func(t *testing.T) {
				for row, err := range goe.Select[struct {
					Name      string
					LowerName query.Function[string]
				}](&struct {
					Name      *string
					LowerName *query.Function[string]
				}{
					Name:      &db.Animal.Name,
					LowerName: function.ToLower(&db.Animal.Name),
				}).Rows() {
					if err != nil {
						t.Fatalf("Expected select, got error: %v", err)
					}
					if strings.ToLower(row.Name) != row.LowerName.Value {
						t.Fatalf("Expected %v, got: %v", strings.ToLower(row.Name), row.LowerName.Value)
					}
				}
			},
		},
		{
			desc: "Select_Like_ToUpper",
			testCase: func(t *testing.T) {
				var a []Animal
				a, err = goe.List(db.Animal).Where(where.Like(function.ToUpper(&db.Animal.Name), "%CAT%")).AsSlice()
				if err != nil {
					t.Fatalf("Expected select, got error: %v", err)
				}
				if len(a) != 3 {
					t.Errorf("Expected 3, got %v", len(a))
				}
			},
		},
		{
			desc: "Select_Equals_ToUpper",
			testCase: func(t *testing.T) {
				var a []Animal
				a, err = goe.List(db.Animal).Where(where.Equals(function.ToUpper(&db.Animal.Name), function.Argument("CAT"))).AsSlice()
				if err != nil {
					t.Fatalf("Expected select, got error: %v", err)
				}
				if strings.ToUpper(a[0].Name) != "CAT" {
					t.Errorf("Expected CAT, got %v", strings.ToUpper(a[0].Name))
				}
			},
		},
		{
			desc: "List_Filter_Order",
			testCase: func(t *testing.T) {
				var a []Animal
				var s []string
				a, err = goe.List(db.Animal).
					Filter(
						where.And(where.In(&db.Animal.Name, s),
							where.And(where.Equals(&db.Animal.Id, 0),
								where.And(
									where.Equals(&db.Animal.Name, ""),
									where.Like(&db.Animal.Name, "%o%"),
								),
							),
						),
					).
					OrderByDesc(&db.Animal.Id).AsSlice()
				if err != nil {
					t.Fatalf("Expected List, got error: %v", err)
				}
				if a[0].Id != animals[len(animals)-1].Id {
					t.Errorf("Expected %v, got %v", animals[len(animals)-1].Id, a[0].Id)
				}
			},
		},
		{
			desc: "Select_Count",
			testCase: func(t *testing.T) {
				a := runSelect(t, goe.Select[struct {
					query.Count
				}](&struct {
					*query.Count
				}{
					aggregate.Count(&db.Animal.Id),
				}).Rows())

				if int(a[0].Value) != len(animals) {
					t.Errorf("Expected %v got: %v", len(animals), a[0].Value)
				}
			},
		},
		{
			desc: "Select_Max_Min",
			testCase: func(t *testing.T) {
				a := runSelect(t, goe.Select[struct {
					Max query.Max
					Min query.Min
				}](&struct {
					*query.Max
					*query.Min
				}{
					aggregate.Max(&db.Exam.Score),
					aggregate.Min(&db.Exam.Score),
				}).Rows())

				if a[0].Min.Value != float64(exams[1].Score) {
					t.Errorf("Expected %v got: %v", float64(exams[1].Score), a[0].Min.Value)
				}

				if a[0].Max.Value != float64(exams[0].Score) {
					t.Errorf("Expected %v got: %v", float64(exams[0].Score), a[0].Max.Value)
				}
			},
		},
		{
			desc: "Select_Sum",
			testCase: func(t *testing.T) {
				a := runSelect(t, goe.Select[struct {
					query.Sum
				}](&struct {
					*query.Sum
				}{
					aggregate.Sum(&db.Exam.Minimum),
				}).Rows())

				if a[0].Value != 16.5 {
					t.Errorf("Expected %v got: %v", float64(exams[1].Score), a[0].Value)
				}
			},
		},
		{
			desc: "Select_Avg",
			testCase: func(t *testing.T) {
				a := runSelect(t, goe.Select[struct {
					query.Avg
				}](&struct {
					*query.Avg
				}{
					aggregate.Avg(&db.Exam.Minimum),
				}).Rows())

				if a[0].Value != float64(exams[0].Minimum) {
					t.Errorf("Expected %v got: %v", float64(exams[0].Minimum), a[0].Value)
				}
			},
		},
		{
			desc: "Select_Filter",
			testCase: func(t *testing.T) {
				a, err := goe.Select[struct {
					Id int
				}](&struct {
					Id *int
				}{
					&db.Animal.Id,
				}).Filter(where.Equals(&db.Animal.Id, 0)).AsSlice()

				if err != nil {
					t.Errorf("Expected filter, got error: %v", err)
				}

				if len(a) != len(animals) {
					t.Errorf("Expected %v got: %v", len(animals), len(a))
				}

				a, err = goe.Select[struct {
					Id int
				}](&struct {
					Id *int
				}{
					&db.Animal.Id,
				}).Where(where.Less(&db.Animal.Id, animals[3].Id)).Filter(where.LessEquals(&db.Animal.Id, 0)).AsSlice()

				if err != nil {
					t.Errorf("Expected filter, got error: %v", err)
				}

				if len(a) != 3 {
					t.Errorf("Expected %v got: %v", 3, len(a))
				}

				a, err = goe.Select[struct {
					Id int
				}](&struct {
					Id *int
				}{
					&db.Animal.Id,
				}).Where(where.And(
					where.Less(&db.Animal.Id, animals[3].Id),
					where.LessEquals(&db.Animal.Id, animals[2].Id),
				)).Filter(where.LessEquals(&db.Animal.Id, animals[2].Id)).AsSlice()

				if err != nil {
					t.Errorf("Expected filter, got error: %v", err)
				}

				if len(a) != 3 {
					t.Errorf("Expected %v got: %v", 3, len(a))
				}

				a, err = goe.Select[struct {
					Id int
				}](&struct {
					Id *int
				}{
					&db.Animal.Id,
				}).Where(where.And(
					where.Less(&db.Animal.Id, animals[3].Id),
					where.LessEquals(&db.Animal.Id, animals[2].Id),
				)).Filter(where.In(&db.Animal.Id, []int{animals[0].Id, animals[2].Id, animals[1].Id})).AsSlice()

				if err != nil {
					t.Errorf("Expected filter, got error: %v", err)
				}

				if len(a) != 3 {
					t.Errorf("Expected %v got: %v", 3, len(a))
				}

				var in []int
				a, err = goe.Select[struct {
					Id int
				}](&struct {
					Id *int
				}{
					&db.Animal.Id,
				}).Where(where.And(
					where.Less(&db.Animal.Id, animals[3].Id),
					where.LessEquals(&db.Animal.Id, animals[2].Id),
				)).Filter(where.In(&db.Animal.Id, in)).AsSlice()

				if err != nil {
					t.Errorf("Expected filter, got error: %v", err)
				}

				if len(a) != 3 {
					t.Errorf("Expected %v got: %v", 3, len(a))
				}
			},
		},
		{
			desc: "Select_Where_In_Slice",
			testCase: func(t *testing.T) {
				a, err := goe.List(db.Animal).Where(
					where.And(where.LessEquals(&db.Animal.Id, animals[2].Id), where.In(&db.Animal.Name, []string{"Cat", "Dog"})),
				).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select where in, got error: %v", err)
				}

				if len(a) != 2 {
					t.Errorf("Expected 2, got %v", len(a))
				}
			},
		},
		{
			desc: "Select_Where_Not_In_Slice",
			testCase: func(t *testing.T) {
				a, err := goe.List(db.Animal).Where(where.NotIn(&db.Animal.Name, []string{"Cat"})).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select where in, got error: %v", err)
				}

				if len(a) != (len(animals) - 1) {
					t.Errorf("Expected %v, got %v", (len(animals) - 1), len(a))
				}
			},
		},
		{
			desc: "Select_Where_In_Query_No_Values",
			testCase: func(t *testing.T) {
				querySelect := goe.Select[any](&struct{ Name *string }{Name: &db.Animal.Name}).AsQuery()

				//whereIn as the first where argument
				a, err := goe.List(db.Animal).Where(
					where.And(
						where.In(&db.Animal.Name, querySelect),
						where.LessEquals(&db.Animal.Id, animals[1].Id),
					),
				).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select where in, got error: %v", err)
				}

				if len(a) != 2 {
					t.Errorf("Expected 2, got %v", len(a))
				}
			},
		},
		{
			desc: "Select_Where_NotIn_Query_No_Values",
			testCase: func(t *testing.T) {
				querySelect := goe.Select[any](&struct{ Name *string }{Name: &db.Animal.Name}).AsQuery()
				if err != nil {
					t.Fatalf("Expected a query, got error: %v", err)
				}

				//whereIn as the first where argument
				a, err := goe.List(db.Animal).Where(
					where.And(
						where.NotIn(&db.Animal.Name, querySelect),
						where.LessEquals(&db.Animal.Id, animals[1].Id),
					),
				).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select where in, got error: %v", err)
				}

				if len(a) != 0 {
					t.Errorf("Expected 0, got %v", len(a))
				}
			},
		},
		{
			desc: "Select_Where_In_Query",
			testCase: func(t *testing.T) {
				querySelect := goe.Select[any](&struct{ Name *string }{Name: &db.Animal.Name}).
					Joins(
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
						join.Join[uuid.UUID](&db.AnimalFood.FoodId, &db.Food.Id)).
					Where(
						where.Or(where.Equals(&db.Animal.Name, "Cat"), where.In(&db.Food.Name, []string{foods[0].Name, foods[1].Name}))).
					AsQuery()
				if err != nil {
					t.Fatalf("Expected a query, got error: %v", err)
				}

				//whereIn as the first where argument
				a, err := goe.List(db.Animal).Where(
					where.And(where.In(&db.Animal.Name, querySelect), where.LessEquals(&db.Animal.Id, animals[1].Id)),
				).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select where in, got error: %v", err)
				}

				if len(a) != 2 {
					t.Errorf("Expected 2, got %v", len(a))
				}

				//whereIn as the last where argument
				a, err = goe.List(db.Animal).Where(
					where.And(
						where.LessEquals(&db.Animal.Id, animals[1].Id),
						where.In(&db.Animal.Name, querySelect),
					)).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select where in, got error: %v", err)
				}
				if len(a) != 2 {
					t.Errorf("Expected 2, got %v", len(a))
				}

				//whereIn as the middle where argument
				a, err = goe.List(db.Animal).Where(
					where.And(
						where.And(
							where.NotEquals(&db.Animal.Id, animals[2].Id),
							where.In(&db.Animal.Name, querySelect)),
						where.LessEquals(&db.Animal.Id, animals[2].Id),
					)).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select where in, got error: %v", err)
				}

				if len(a) != 2 {
					t.Errorf("Expected 2, got %v", len(a))
				}
			},
		},
		{
			desc: "Select_Where_NotIn_Query",
			testCase: func(t *testing.T) {
				querySelect := goe.Select[any](&struct{ Name *string }{Name: &db.Animal.Name}).
					Joins(
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
						join.Join[uuid.UUID](&db.AnimalFood.FoodId, &db.Food.Id)).
					Where(where.Or(
						where.Equals(&db.Animal.Name, "Cat"),
						where.NotIn(&db.Food.Name, []string{foods[0].Name, foods[1].Name}),
					)).
					AsQuery()
				if err != nil {
					t.Fatalf("Expected a query, got error: %v", err)
				}

				//whereIn as the first where argument
				a, err := goe.List(db.Animal).Where(
					where.And(
						where.NotIn(&db.Animal.Name, querySelect),
						where.LessEquals(&db.Animal.Id, animals[1].Id),
					)).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select where in, got error: %v", err)
				}

				if len(a) != 1 {
					t.Errorf("Expected 1, got %v", len(a))
				}

				//whereIn as the last where argument
				a, err = goe.List(db.Animal).Where(
					where.And(
						where.LessEquals(&db.Animal.Id, animals[1].Id),
						where.NotIn(&db.Animal.Name, querySelect),
					)).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select where in, got error: %v", err)
				}
				if len(a) != 1 {
					t.Errorf("Expected 1, got %v", len(a))
				}

				//whereIn as the middle where argument
				a, err = goe.List(db.Animal).Where(
					where.And(
						where.And(
							where.NotEquals(&db.Animal.Id, animals[2].Id),
							where.NotIn(&db.Animal.Name, querySelect),
						),
						where.LessEquals(&db.Animal.Id, animals[2].Id),
					)).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select where in, got error: %v", err)
				}

				if len(a) != 1 {
					t.Errorf("Expected 1, got %v", len(a))
				}
			},
		},
		{
			desc: "List_As_Pagination",
			testCase: func(t *testing.T) {
				var p *goe.Pagination[Animal]
				p, err = goe.List(db.Animal).AsPagination(1, 10)
				if err != nil {
					t.Fatalf("Expected pagination, got: %v", err)
				}

				if p.TotalValues != int64(len(animals)) {
					t.Errorf("Expected %v, got %v", len(animals), p.TotalValues)
				}

				if p.TotalPages != 4 {
					t.Errorf("Expected 4, got %v", p.TotalPages)
				}

				if p.PageSize != 10 {
					t.Errorf("Expected 10, got %v", p.PageSize)
				}

				if p.CurrentPage != 1 {
					t.Errorf("Expected 1, got %v", p.CurrentPage)
				}

				if p.NextPage != 2 {
					t.Errorf("Expected 2, got %v", p.NextPage)
				}

				if p.PreviousPage != 1 {
					t.Errorf("Expected 1, got %v", p.PreviousPage)
				}

				if p.HasNextPage != true {
					t.Errorf("Expected true, got %v", p.HasNextPage)
				}

				if p.HasPreviousPage != false {
					t.Errorf("Expected false, got %v", p.HasPreviousPage)
				}

				if p.PageValues != 10 {
					t.Errorf("Expected 10, got %v", p.PageValues)
				}

				if p.StartIndex != 1 {
					t.Errorf("Expected 1, got %v", p.StartIndex)
				}

				if p.EndIndex != 10 {
					t.Errorf("Expected 10, got %v", p.EndIndex)
				}

				//navigate to second page
				p, err = goe.List(db.Animal).AsPagination(p.NextPage, p.PageSize)
				if err != nil {
					t.Fatalf("Expected pagination, got: %v", err)
				}

				if p.StartIndex != 11 {
					t.Errorf("Expected 1, got %v", p.StartIndex)
				}

				if p.EndIndex != 20 {
					t.Errorf("Expected 10, got %v", p.EndIndex)
				}

				//navigate to third page
				p, err = goe.List(db.Animal).AsPagination(p.NextPage, p.PageSize)
				if err != nil {
					t.Fatalf("Expected pagination, got: %v", err)
				}

				if p.StartIndex != 21 {
					t.Errorf("Expected 1, got %v", p.StartIndex)
				}

				if p.EndIndex != 30 {
					t.Errorf("Expected 10, got %v", p.EndIndex)
				}

				p, err = goe.List(db.Animal).AsPagination(p.TotalPages, p.PageSize)
				if err != nil {
					t.Fatalf("Expected pagination, got: %v", err)
				}

				if p.TotalValues != int64(len(animals)) {
					t.Errorf("Expected %v, got %v", len(animals), p.TotalValues)
				}

				if p.TotalPages != 4 {
					t.Errorf("Expected 4, got %v", p.TotalPages)
				}

				if p.PageSize != 10 {
					t.Errorf("Expected 10, got %v", p.PageSize)
				}

				if p.CurrentPage != 4 {
					t.Errorf("Expected 4, got %v", p.CurrentPage)
				}

				if p.NextPage != 4 {
					t.Errorf("Expected 4, got %v", p.NextPage)
				}

				if p.PreviousPage != 3 {
					t.Errorf("Expected 3, got %v", p.PreviousPage)
				}

				if p.HasNextPage != false {
					t.Errorf("Expected false, got %v", p.HasNextPage)
				}

				if p.HasPreviousPage != true {
					t.Errorf("Expected true, got %v", p.HasPreviousPage)
				}

				if p.PageValues != 2 {
					t.Errorf("Expected 2, got %v", p.PageValues)
				}

				if p.StartIndex != 31 {
					t.Errorf("Expected 31, got %v", p.StartIndex)
				}

				if p.EndIndex != 32 {
					t.Errorf("Expected 31, got %v", p.EndIndex)
				}
			},
		},
		{
			desc: "List_As_Pagination_Page_And_Size_0",
			testCase: func(t *testing.T) {
				var p *goe.Pagination[Animal]
				p, err = goe.List(db.Animal).AsPagination(0, 1)
				if err != nil {
					t.Fatalf("Expected pagination, got: %v", err)
				}
				if p.CurrentPage != 1 {
					t.Fatalf("Expected page 1, got: %v", p.CurrentPage)
				}

				p, err = goe.List(db.Animal).AsPagination(1, 0)
				if err != nil {
					t.Fatalf("Expected pagination, got: %v", err)
				}
				if p.PageSize != 10 {
					t.Fatalf("Expected size 10, got: %v", p.PageSize)
				}
			},
		},
		{
			desc: "Select_As_Pagination",
			testCase: func(t *testing.T) {
				var p *goe.Pagination[Animal]
				p, err = goe.List(db.Animal).
					Joins(join.Join[uuid.UUID](&db.Animal.HabitatId, &db.Habitat.Id)).
					Where(where.Like(&db.Animal.Name, "%at")).
					OrderByDesc(&db.Animal.Id).
					AsPagination(1, 10)
				if err != nil {
					t.Fatalf("Expected pagination, got: %v", err)
				}

				if p.TotalValues != 3 {
					t.Errorf("Expected 3, got %v", p.TotalValues)
				}
			},
		},
		{
			desc: "Find_ByValue",
			testCase: func(t *testing.T) {
				var a *Animal
				a, err = goe.Find(db.Animal).ByValue(Animal{Name: "Cat"})
				if err != nil {
					t.Fatalf("Expected find, got: %v", err)
				}
				if a.Name != "Cat" {
					t.Fatalf("Expected Cat, got: %v", a.Name)
				}

				customErr := errors.New("my custom error")

				_, err = goe.Find(db.Animal).ByValue(Animal{Id: -1})
				if !errors.Is(err, goe.ErrNotFound) {
					t.Fatalf("Expected goe.ErrNotFound, got: %v", err)
				}

				_, err = goe.Find(db.Animal).OnErrNotFound(customErr).ByValue(Animal{Id: -1})
				if !errors.Is(err, customErr) {
					t.Fatalf("Expected customErr, got: %v", err)
				}
			},
		},
		{
			desc: "Select_As_Pagination_Anonymous",
			testCase: func(t *testing.T) {
				var p *goe.Pagination[struct {
					Animal    string
					Habitat   *string
					HabitatId *uuid.UUID
				}]
				p, err = goe.Select[struct {
					Animal    string
					Habitat   *string
					HabitatId *uuid.UUID
				}](&struct {
					Animal    *string
					Habitat   *string
					HabitatId *uuid.UUID
				}{
					Animal:    &db.Animal.Name,
					Habitat:   &db.Habitat.Name,
					HabitatId: &db.Habitat.Id,
				}).
					Joins(join.LeftJoin[uuid.UUID](&db.Animal.HabitatId, &db.Habitat.Id)).
					OrderByAsc(&db.Animal.Id).
					AsPagination(1, 10)
				if err != nil {
					t.Fatalf("Expected pagination, got: %v", err)
				}

				if p.TotalValues != 32 {
					t.Errorf("Expected 32, got %v", p.TotalValues)
				}
			},
		},
		{
			desc: "Find",
			testCase: func(t *testing.T) {
				var a *Animal
				a, err = goe.Find(db.Animal).ById(Animal{Id: animals[0].Id})
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
				a, err = goe.Find(db.AnimalFood).ById(AnimalFood{AnimalId: animals[0].Id, FoodId: foods[0].Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}
				if a.AnimalId != animals[0].Id {
					t.Errorf("Expected a %v, got %v", animals[0].Id, a.AnimalId)
				}
				if a.FoodId != foods[0].Id {
					t.Errorf("Expected a %v, got %v", foods[0].Id, a.FoodId)
				}
			},
		},
		{
			desc: "Select_Where_Greater",
			testCase: func(t *testing.T) {
				e := runSelect(t, goe.List(db.Exam).
					Where(where.GreaterArg[float32](&db.Exam.Score, &db.Exam.Minimum)).Rows())
				if len(e) != 1 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}

				e = nil
				e = runSelect(t, goe.List(db.Exam).
					Where(where.Greater(&db.Exam.Score, float32(5.5))).Rows())
				if len(e) != 1 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}
			},
		},
		{
			desc: "Select_Where_GreaterEquals",
			testCase: func(t *testing.T) {
				e := runSelect(t, goe.List(db.Exam).
					Where(where.GreaterEqualsArg[float32](&db.Exam.Score, &db.Exam.Minimum)).Rows())
				if len(e) != 2 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}

				e = nil
				e = runSelect(t, goe.List(db.Exam).
					Where(where.GreaterEquals(&db.Exam.Score, float32(5.5))).Rows())
				if len(e) != 2 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}
			},
		},
		{
			desc: "Select_Where_Less",
			testCase: func(t *testing.T) {
				e := runSelect(t, goe.List(db.Exam).
					Where(where.LessArg[float32](&db.Exam.Score, &db.Exam.Minimum)).Rows())
				if len(e) != 1 {
					t.Errorf("Expected %v, got %v", 1, len(e))
				}

				e = nil
				e = runSelect(t, goe.List(db.Exam).
					Where(where.Less(&db.Exam.Score, float32(5.5))).Rows())
				if len(e) != 1 {
					t.Errorf("Expected %v, got %v", 1, len(e))
				}
			},
		},
		{
			desc: "Select_Where_LessEquals",
			testCase: func(t *testing.T) {
				e := runSelect(t, goe.List(db.Exam).
					Where(where.LessEqualsArg[float32](&db.Exam.Score, &db.Exam.Minimum)).Rows())
				if len(e) != 2 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}

				e = nil
				e = runSelect(t, goe.List(db.Exam).
					Where(where.LessEquals(&db.Exam.Score, float32(5.5))).Rows())
				if len(e) != 2 {
					t.Errorf("Expected a %v, got %v", 1, len(e))
				}
			},
		},
		{
			desc: "Select_Where_Like",
			testCase: func(t *testing.T) {
				a := runSelect(t, goe.List(db.Animal).
					Where(where.Like(&db.Animal.Name, "%at%")).Rows())
				if len(a) != 3 {
					t.Errorf("Expected %v animals, got %v", 3, len(a))
				}
			},
		},
		{
			desc: "Select_Where_NotLike",
			testCase: func(t *testing.T) {
				a := runSelect(t, goe.List(db.Animal).
					Where(where.NotLike(&db.Animal.Name, "%at%")).Rows())
				if len(a) != len(animals)-3 {
					t.Errorf("Expected %v animals, got %v", len(animals)-3, len(a))
				}
			},
		},
		{
			desc: "Select_Where_Equals_Nil",
			testCase: func(t *testing.T) {
				qr := goe.List(db.Animal).Where(where.Equals[*uuid.UUID](&db.Animal.HabitatId, nil)).Rows()
				a := runSelect(t, qr)
				if len(a) != 24 {
					t.Errorf("Expected %v animals, got %v", 22, len(a))
				}
			},
		},
		{
			desc: "Select_Where_NotEquals_Nil",
			testCase: func(t *testing.T) {
				var bb *[]byte
				qr := goe.List(db.Animal).Where(where.NotEquals(&db.Animal.InfoId, bb)).Rows()
				a := runSelect(t, qr)
				if len(a) != len(infos) {
					t.Errorf("Expected %v animals, got %v", len(infos), len(a))
				}
			},
		},
		{
			desc: "Find_Not_Found",
			testCase: func(t *testing.T) {
				_, err = goe.Find(db.Animal).ById(Animal{Id: -1})
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}
			},
		},
		{
			desc: "Select_Order_By_Asc",
			testCase: func(t *testing.T) {
				qr := goe.List(db.Animal).OrderByAsc(&db.Animal.Id).Rows()
				a := runSelect(t, qr)
				if a[0].Id > a[1].Id {
					t.Errorf("Expected animals order by asc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Order_By_Desc",
			testCase: func(t *testing.T) {
				qr := goe.List(db.Animal).OrderByDesc(&db.Animal.Id).Rows()
				a := runSelect(t, qr)
				if a[0].Id < a[1].Id {
					t.Errorf("Expected animals order by desc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Join",
			testCase: func(t *testing.T) {
				qr := goe.List(db.Animal).
					Joins(
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
						join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
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
				qr := goe.List(db.Animal).
					Where(
						where.And(
							where.EqualsArg[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
							where.EqualsArg[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
						),
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
			desc: "Select_Join_Where",
			testCase: func(t *testing.T) {
				qr := goe.List(db.Food).
					Joins(
						join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
					).
					Where(
						where.Equals(&db.Animal.Name, animals[0].Name)).Rows()
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
				qr := goe.List(db.Animal).
					Joins(
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
						join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
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
				qr := goe.List(db.Animal).
					Joins(
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
						join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
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
				qr := goe.List(db.Animal).
					Joins(
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
						join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
					).
					Where(
						where.Equals(&db.Food.Id, foods[0].Id),
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
				qr := goe.List(db.Animal).
					Joins(
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
						join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
					).
					Where(
						where.Equals(&db.Food.Id, foods[0].Id),
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
				qr := goe.List(db.Food).
					Joins(
						join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
						join.Join[uuid.UUID](&db.Animal.HabitatId, &db.Habitat.Id),
					).
					Where(where.Equals(&db.Habitat.Id, habitats[0].Id)).Rows()
				f := runSelect(t, qr)

				if len(f) != 2 {
					t.Errorf("Expected 2, got : %v", len(f))
				}
			},
		},
		{
			desc: "Select_Join_One_To_One",
			testCase: func(t *testing.T) {
				qr := goe.List(db.Animal).
					Joins(
						join.Join[[]byte](&db.Animal.InfoId, &db.Info.Id),
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
				qr := goe.List(db.Info).
					Joins(
						join.Join[int](&db.Status.Id, &db.Info.StatusId),
						join.Join[[]byte](&db.Animal.InfoId, &db.Info.Id),
						join.Join[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
						join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
					).
					Where(where.Equals(&db.Food.Id, foods[0].Id)).Rows()
				s := runSelect(t, qr)

				if len(s) != 2 {
					t.Errorf("Expected 2, got : %v", len(s))
				}
			},
		},
		{
			desc: "Select_Join_Name",
			testCase: func(t *testing.T) {
				qr := goe.List(db.Habitat).
					Joins(
						join.Join[string](&db.Habitat.Name, &db.Weather.Name),
					).Rows()
				h := runSelect(t, qr)

				if h[0].Name != "Ocean" {
					t.Errorf("Expected Ocean, got : %v", h[0].Name)
				}
			},
		},
		{
			desc: "Select_Benchmark_Joins",
			testCase: func(t *testing.T) {
				for _, err := range goe.List(db.Food).
					Joins(
						join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
						join.Join[int](&db.AnimalFood.AnimalId, &db.Animal.Id),
						join.Join[uuid.UUID](&db.Animal.HabitatId, &db.Habitat.Id),
						join.Join[int](&db.Habitat.WeatherId, &db.Weather.Id),
					).
					Where(
						where.And(
							where.Equals(&db.Food.Id, foods[0].Id),
							where.Equals(&db.Food.Name, foods[0].Name),
						),
					).
					Rows() {
					if err != nil {
						t.Fatal(err)
					}
				}
			},
		},
		{
			desc: "Select_User_And_Roles",
			testCase: func(t *testing.T) {
				var q []struct {
					User    string
					Role    *string
					EndTime *time.Time
				}

				for row, err := range goe.Select[struct {
					User    string
					Role    *string
					EndTime *time.Time
				}](&struct {
					User    *string
					Role    *string
					EndTime **time.Time
				}{
					User:    &db.User.Name,
					Role:    &db.Role.Name,
					EndTime: &db.UserRole.EndDate,
				}).
					Joins(
						join.LeftJoin[int](&db.User.Id, &db.UserRole.UserId),
						join.LeftJoin[int](&db.UserRole.RoleId, &db.Role.Id),
					).
					OrderByAsc(&db.User.Id).Rows() {

					if err != nil {
						t.Fatal(err)
					}

					q = append(q, row)
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
					User    *string
					Role    string
					EndTime *time.Time
				}

				for row, err := range goe.Select[struct {
					User    *string
					Role    string
					EndTime *time.Time
				}](&struct {
					User    *string
					Role    *string
					EndTime **time.Time
				}{
					User:    &db.User.Name,
					Role:    &db.Role.Name,
					EndTime: &db.UserRole.EndDate,
				}).
					Joins(
						join.RightJoin[int](&db.UserRole.UserId, &db.User.Id),
						join.RightJoin[int](&db.Role.Id, &db.UserRole.RoleId),
					).
					OrderByAsc(&db.Role.Id).Rows() {

					if err != nil {
						t.Fatal(err)
					}

					q = append(q, row)
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

				for row, err := range goe.Select[struct {
					JobTitle string
					Person   string
				}](&struct {
					JobTitle *string
					Person   *string
				}{
					JobTitle: &db.JobTitle.Name,
					Person:   &db.Person.Name,
				}).
					Joins(
						join.Join[int](&db.Person.Id, &db.PersonJobTitle.PersonId),
						join.Join[int](&db.PersonJobTitle.JobTitleId, &db.JobTitle.Id),
					).Rows() {

					if err != nil {
						t.Fatal(err)
					}

					pj = append(pj, row)
				}

				if len(pj) != len(personJobs) {
					t.Errorf("Expected %v, got : %v", len(personJobs), len(pj))
				}
			},
		},
		{
			desc: "Select_Context_Cancel",
			testCase: func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				for _, err := range goe.ListContext(ctx, db.Animal).Rows() {
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
				for _, err := range goe.ListContext(ctx, db.Animal).Rows() {
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Errorf("Expected a context.DeadlineExceeded, got error: %v", err)
					}
				}
			},
		},
		{
			desc: "Select_SameQuery_Race",
			testCase: func(t *testing.T) {
				var wg sync.WaitGroup
				selectQuery := goe.List(db.Animal)
				for range 2 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						selectQuery.Joins(join.Join[int](&db.AnimalFood.AnimalId, &db.Animal.Id)).AsSlice()
						selectQuery.
							Where(
								where.And(
									where.EqualsArg[int](&db.Animal.Id, &db.AnimalFood.AnimalId),
									where.EqualsArg[uuid.UUID](&db.Food.Id, &db.AnimalFood.FoodId),
								),
							).AsSlice()
					}()
				}
				wg.Wait()
			},
		},
	}
	for _, tC := range testCases {
		if tC.desc == "List_Filter_Order" {
			t.Run(tC.desc, tC.testCase)
		}
	}
}

// TODO: remove this
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
