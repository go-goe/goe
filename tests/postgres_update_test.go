package tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olauro/goe"
	"github.com/olauro/goe/jn"
	"github.com/olauro/goe/wh"
)

func TestPostgresUpdate(t *testing.T) {
	db, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Update_Flag",
			testCase: func(t *testing.T) {
				f := Flag{
					Id:      uuid.New(),
					Name:    "Flag",
					Float32: 1.1,
					Float64: 2.2,
					Today:   time.Now(),
					Int:     -1,
					Int8:    -8,
					Int16:   -16,
					Int32:   -32,
					Int64:   -64,
					Uint:    1,
					Uint8:   8,
					Uint16:  16,
					Uint32:  32,
					Bool:    true,
				}
				err = db.Insert(db.Flag).Value(&f)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}

				ff := Flag{
					Name:    "Flag_Test",
					Float32: 3.3,
					Float64: 4.4,
					Bool:    false,
				}
				err = db.Update(&db.Flag.Name, &db.Flag.Bool, &db.Flag.Float64, &db.Flag.Float32).Where(wh.Equals(&db.Flag.Id, f.Id)).Value(ff)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				var fselect Flag
				err = db.Select(db.Flag).From(db.Flag).Where(wh.Equals(&db.Flag.Id, f.Id)).Scan(&fselect)

				if fselect.Name != ff.Name {
					t.Errorf("Expected a update on name, got : %v", fselect.Name)
				}
				if fselect.Float32 != ff.Float32 {
					t.Errorf("Expected a update on float32, got : %v", fselect.Float32)
				}
				if fselect.Float64 != ff.Float64 {
					t.Errorf("Expected a update on float32, got : %v", fselect.Float64)
				}
				if fselect.Bool != ff.Bool {
					t.Errorf("Expected a update on float32, got : %v", fselect.Bool)
				}
			},
		},
		{
			desc: "Update_Flag_Pointer",
			testCase: func(t *testing.T) {
				f := Flag{
					Id:      uuid.New(),
					Name:    "Flag",
					Float32: 1.1,
					Float64: 2.2,
					Today:   time.Now(),
					Int:     -1,
					Int8:    -8,
					Int16:   -16,
					Int32:   -32,
					Int64:   -64,
					Uint:    1,
					Uint8:   8,
					Uint16:  16,
					Uint32:  32,
					Bool:    true,
				}
				err = db.Insert(db.Flag).Value(&f)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}

				ff := Flag{
					Name:    "Flag_Test",
					Float32: 3.3,
					Float64: 4.4,
					Bool:    false,
				}
				err = db.Update(&db.Flag.Name, &db.Flag.Bool, &db.Flag.Float64, &db.Flag.Float32).Where(wh.Equals(&db.Flag.Id, f.Id)).Value(&ff)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				var fselect Flag
				err = db.Select(db.Flag).From(db.Flag).Where(wh.Equals(&db.Flag.Id, f.Id)).Scan(&fselect)

				if fselect.Name != ff.Name {
					t.Errorf("Expected a update on name, got : %v", fselect.Name)
				}
				if fselect.Float32 != ff.Float32 {
					t.Errorf("Expected a update on float32, got : %v", fselect.Float32)
				}
				if fselect.Float64 != ff.Float64 {
					t.Errorf("Expected a update on float32, got : %v", fselect.Float64)
				}
				if fselect.Bool != ff.Bool {
					t.Errorf("Expected a update on float32, got : %v", fselect.Bool)
				}
			},
		},
		{
			desc: "Update_Animal",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				w := Weather{
					Name: "Warm",
				}
				err = db.Insert(db.Weather).Value(&w)
				if err != nil {
					t.Errorf("Expected a insert weather, got error: %v", err)
				}

				h := Habitat{
					Id:        uuid.New(),
					Name:      "City",
					IdWeather: w.Id,
				}
				err = db.Insert(db.Habitat).Value(&h)
				if err != nil {
					t.Errorf("Expected a insert habitat, got error: %v", err)
				}

				a.IdHabitat = &h.Id
				a.Name = "Update Cat"
				err = db.Update(&db.Animal.IdHabitat, &db.Animal.Name).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(a)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				var aselect Animal
				err = db.Select(db.Animal).From(db.Animal).Where(wh.Equals(&db.Animal.Id, a.Id)).Scan(&aselect)

				if aselect.IdHabitat == nil || *aselect.IdHabitat != h.Id {
					t.Errorf("Expected a update on id habitat, got : %v", aselect.IdHabitat)
				}
				if aselect.Name != "Update Cat" {
					t.Errorf("Expected a update on name, got : %v", aselect.Name)
				}
			},
		},
		{
			desc: "Update_Animal_Pointer",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				w := Weather{
					Name: "Warm",
				}
				err = db.Insert(db.Weather).Value(&w)
				if err != nil {
					t.Errorf("Expected a insert weather, got error: %v", err)
				}

				h := Habitat{
					Id:        uuid.New(),
					Name:      "City",
					IdWeather: w.Id,
				}
				err = db.Insert(db.Habitat).Value(&h)
				if err != nil {
					t.Errorf("Expected a insert habitat, got error: %v", err)
				}

				a.IdHabitat = &h.Id
				a.Name = "Update Cat"
				err = db.Update(&db.Animal.IdHabitat, &db.Animal.Name).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(&a)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				var aselect Animal
				err = db.Select(db.Animal).From(db.Animal).Where(wh.Equals(&db.Animal.Id, a.Id)).Scan(&aselect)

				if aselect.IdHabitat == nil || *aselect.IdHabitat != h.Id {
					t.Errorf("Expected a update on id habitat, got : %v", aselect.IdHabitat)
				}
				if aselect.Name != "Update Cat" {
					t.Errorf("Expected a update on name, got : %v", aselect.Name)
				}
			},
		},
		{
			desc: "Update_Animal_All_Fields",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				w := Weather{
					Name: "Warm",
				}
				err = db.Insert(db.Weather).Value(&w)
				if err != nil {
					t.Errorf("Expected a insert weather, got error: %v", err)
				}

				h := Habitat{
					Id:        uuid.New(),
					Name:      "City",
					IdWeather: w.Id,
				}
				err = db.Insert(db.Habitat).Value(&h)
				if err != nil {
					t.Errorf("Expected a insert habitat, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				a.IdHabitat = &h.Id
				a.Name = "Update Cat"
				err = db.Update(db.Animal).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(a)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				var aselect Animal
				err = db.Select(db.Animal).From(db.Animal).Where(wh.Equals(&db.Animal.Id, a.Id)).Scan(&aselect)

				if aselect.IdHabitat == nil || *aselect.IdHabitat != h.Id {
					t.Errorf("Expected a update on id habitat, got : %v", aselect.IdHabitat)
				}
				if aselect.Name != "Update Cat" {
					t.Errorf("Expected a update on name, got : %v", aselect.Name)
				}
			},
		},
		{
			desc: "Update_PersonJobs",
			testCase: func(t *testing.T) {
				persons := []Person{
					{Name: "Jhon"},
					{Name: "Laura"},
					{Name: "Luana"},
				}
				err = db.Insert(db.Person).Value(&persons)
				if err != nil {
					t.Fatalf("Expected insert persons, got error: %v", err)
				}

				jobs := []JobTitle{
					{Name: "Developer"},
					{Name: "Designer"},
				}
				err = db.Insert(db.JobTitle).Value(&jobs)
				if err != nil {
					t.Fatalf("Expected insert jobs, got error: %v", err)
				}

				personJobs := []PersonJobTitle{
					{IdPerson: persons[0].Id, IdJobTitle: jobs[0].Id, CreatedAt: time.Now()},
					{IdPerson: persons[1].Id, IdJobTitle: jobs[0].Id, CreatedAt: time.Now()},
					{IdPerson: persons[2].Id, IdJobTitle: jobs[1].Id, CreatedAt: time.Now()},
				}
				err = db.Insert(db.PersonJobTitle).Value(&personJobs)
				if err != nil {
					t.Fatalf("Expected insert personJobs, got error: %v", err)
				}

				pj := []struct {
					JobTitle string
					Person   string
				}{}
				err = db.Select(&db.Person.Name, &db.JobTitle.Name).
					From(db.Person).
					Joins(
						jn.Join[int](&db.Person.Id, &db.PersonJobTitle.IdPerson),
						jn.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(wh.Equals(&db.JobTitle.Id, jobs[0].Id)).Scan(&pj)
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}
				if len(pj) != 2 {
					t.Errorf("Expected %v, got : %v", 2, len(pj))
				}

				err = db.Update(&db.PersonJobTitle.IdJobTitle).Where(
					wh.Equals(&db.PersonJobTitle.IdPerson, persons[2].Id),
					wh.And(),
					wh.Equals(&db.PersonJobTitle.IdJobTitle, jobs[1].Id),
				).Value(PersonJobTitle{IdJobTitle: jobs[0].Id})
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				err = db.Select(&db.Person.Name, &db.JobTitle.Name).
					From(db.Person).
					Joins(
						jn.Join[int](&db.Person.Id, &db.PersonJobTitle.IdPerson),
						jn.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(wh.Equals(&db.JobTitle.Id, jobs[0].Id)).Scan(&pj)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(pj) != 3 {
					t.Errorf("Expected %v, got : %v", 3, len(pj))
				}
			},
		},
		{
			desc: "Update_Invalid_Tables_1",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = db.Update(&db.Animal.IdHabitat, &db.Food.Name).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrTooManyTablesUpdate) {
					t.Errorf("Expected a goe.ErrTooManyTablesUpdate, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Invalid_Tables_2",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = db.Update(db.Animal, db.Food).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrTooManyTablesUpdate) {
					t.Errorf("Expected a goe.ErrTooManyTablesUpdate, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Invalid_Arg",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = db.Update(db.DB).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected a goe.ErrInvalidArg, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Invalid_Where",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = db.Update(db.Animal).Where(wh.Equals(&a.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrInvalidWhere) {
					t.Errorf("Expected a goe.ErrInvalidWhere, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Invalid_Value",
			testCase: func(t *testing.T) {
				a := 1
				err = db.Update(db.Animal).Where(wh.Equals(&db.Animal.Id, 2)).Value(a)
				if !errors.Is(err, goe.ErrInvalidUpdateValue) {
					t.Errorf("Expected a goe.ErrInvalidUpdateValue, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Context_Cancel",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				err = db.UpdateContext(ctx, db.Animal).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, context.Canceled) {
					t.Errorf("Expected a context.Canceled, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Context_Timeout",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond*1)
				defer cancel()
				err = db.UpdateContext(ctx, db.Animal).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("Expected a context.DeadlineExceeded, got error: %v", err)
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
