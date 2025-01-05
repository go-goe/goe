package tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olauro/goe"
	"github.com/olauro/goe/jn"
	"github.com/olauro/goe/query"
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
				err = query.Insert(db.DB, db.Flag).One(&f)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}

				ff := Flag{
					Name:    "Flag_Test",
					Float32: 3.3,
					Float64: 4.4,
					Bool:    false,
				}
				u := query.Update(db.DB, db.Flag).
					Includes(&db.Flag.Name, &db.Flag.Bool).
					Where(wh.Equals(&db.Flag.Id, f.Id))
				err = u.Includes(&db.Flag.Float64, &db.Flag.Float32).Value(ff)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				var fselect *Flag
				fselect, err = query.Find(db.DB, db.Flag, Flag{Id: f.Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

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
				err = query.Insert(db.DB, db.Animal).One(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				w := Weather{
					Name: "Warm",
				}
				err = query.Insert(db.DB, db.Weather).One(&w)
				if err != nil {
					t.Errorf("Expected a insert weather, got error: %v", err)
				}

				h := Habitat{
					Id:        uuid.New(),
					Name:      "City",
					IdWeather: w.Id,
				}
				err = query.Insert(db.DB, db.Habitat).One(&h)
				if err != nil {
					t.Errorf("Expected a insert habitat, got error: %v", err)
				}

				a.IdHabitat = &h.Id
				a.Name = "Update Cat"
				err = query.Save(db.DB, db.Animal).Includes(&db.Animal.IdHabitat).Value(a)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				var aselect *Animal
				aselect, err = query.Find(db.DB, db.Animal, Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

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
				err = query.Insert(db.DB, db.Person).All(persons)
				if err != nil {
					t.Fatalf("Expected insert persons, got error: %v", err)
				}

				jobs := []JobTitle{
					{Name: "Developer"},
					{Name: "Designer"},
				}
				err = query.Insert(db.DB, db.JobTitle).All(jobs)
				if err != nil {
					t.Fatalf("Expected insert jobs, got error: %v", err)
				}

				personJobs := []PersonJobTitle{
					{IdPerson: persons[0].Id, IdJobTitle: jobs[0].Id, CreatedAt: time.Now()},
					{IdPerson: persons[1].Id, IdJobTitle: jobs[0].Id, CreatedAt: time.Now()},
					{IdPerson: persons[2].Id, IdJobTitle: jobs[1].Id, CreatedAt: time.Now()},
				}
				err = query.Insert(db.DB, db.PersonJobTitle).All(personJobs)
				if err != nil {
					t.Fatalf("Expected insert personJobs, got error: %v", err)
				}

				pj := []struct {
					JobTitle string
					Person   string
				}{}
				for row, err := range query.Select(db.DB, &struct {
					JobTitle *string
					Person   *string
				}{
					JobTitle: &db.JobTitle.Name,
					Person:   &db.Person.Name,
				}).From(db.Person).
					Joins(
						jn.Join[int](&db.Person.Id, &db.PersonJobTitle.IdPerson),
						jn.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(wh.Equals(&db.JobTitle.Id, jobs[0].Id)).Rows() {

					if err != nil {
						t.Fatalf("Expected a select, got error: %v", err)
					}
					pj = append(pj, struct {
						JobTitle string
						Person   string
					}{
						JobTitle: query.SafeGet(row.JobTitle),
						Person:   query.SafeGet(row.Person),
					})
				}

				if len(pj) != 2 {
					t.Errorf("Expected %v, got : %v", 2, len(pj))
				}

				err = query.Update(db.DB, db.PersonJobTitle).Includes(&db.PersonJobTitle.IdJobTitle).Where(
					wh.Equals(&db.PersonJobTitle.IdPerson, persons[2].Id),
					wh.And(),
					wh.Equals(&db.PersonJobTitle.IdJobTitle, jobs[1].Id),
				).Value(PersonJobTitle{IdJobTitle: jobs[0].Id})
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				pj = nil
				for row, err := range query.Select(db.DB, &struct {
					JobTitle *string
					Person   *string
				}{
					JobTitle: &db.JobTitle.Name,
					Person:   &db.Person.Name,
				}).From(db.Person).
					Joins(
						jn.Join[int](&db.Person.Id, &db.PersonJobTitle.IdPerson),
						jn.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(wh.Equals(&db.JobTitle.Id, jobs[0].Id)).Rows() {

					if err != nil {
						t.Fatalf("Expected a select, got error: %v", err)
					}
					pj = append(pj, struct {
						JobTitle string
						Person   string
					}{
						JobTitle: query.SafeGet(row.JobTitle),
						Person:   query.SafeGet(row.Person),
					})
				}

				if len(pj) != 3 {
					t.Errorf("Expected %v, got : %v", 3, len(pj))
				}
			},
		},
		{
			desc: "Save_PersonJobs",
			testCase: func(t *testing.T) {
				persons := []Person{
					{Name: "Jhon"},
					{Name: "Laura"},
					{Name: "Luana"},
				}
				err = query.Insert(db.DB, db.Person).All(persons)
				if err != nil {
					t.Fatalf("Expected insert persons, got error: %v", err)
				}

				jobs := []JobTitle{
					{Name: "Developer"},
					{Name: "Designer"},
				}
				err = query.Insert(db.DB, db.JobTitle).All(jobs)
				if err != nil {
					t.Fatalf("Expected insert jobs, got error: %v", err)
				}

				personJobs := []PersonJobTitle{
					{IdPerson: persons[0].Id, IdJobTitle: jobs[0].Id},
					{IdPerson: persons[1].Id, IdJobTitle: jobs[0].Id},
					{IdPerson: persons[2].Id, IdJobTitle: jobs[1].Id},
				}
				err = query.Insert(db.DB, db.PersonJobTitle).All(personJobs)
				if err != nil {
					t.Fatalf("Expected insert personJobs, got error: %v", err)
				}

				pj := []struct {
					JobTitle  string
					Person    string
					CreatedAt time.Time
				}{}
				for row, err := range query.Select(db.DB, &struct {
					JobTitle *string
					Person   *string
				}{
					JobTitle: &db.JobTitle.Name,
					Person:   &db.Person.Name,
				}).From(db.Person).
					Joins(
						jn.Join[int](&db.Person.Id, &db.PersonJobTitle.IdPerson),
						jn.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(wh.Equals(&db.JobTitle.Id, jobs[0].Id)).Rows() {

					if err != nil {
						t.Fatalf("Expected a select, got error: %v", err)
					}
					pj = append(pj, struct {
						JobTitle  string
						Person    string
						CreatedAt time.Time
					}{
						JobTitle: query.SafeGet(row.JobTitle),
						Person:   query.SafeGet(row.Person),
					})
				}

				if len(pj) != 2 {
					t.Errorf("Expected %v, got : %v", 2, len(pj))
				}

				err = query.Save(db.DB, db.PersonJobTitle).Replace(PersonJobTitle{
					IdPerson:   persons[2].Id,
					IdJobTitle: jobs[1].Id}).Value(PersonJobTitle{
					IdJobTitle: jobs[0].Id, CreatedAt: time.Now()})
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				pj = nil
				for row, err := range query.Select(db.DB, &struct {
					JobTitle  *string
					Person    *string
					CreatedAt *time.Time
				}{
					JobTitle:  &db.JobTitle.Name,
					CreatedAt: &db.PersonJobTitle.CreatedAt,
					Person:    &db.Person.Name,
				}).From(db.Person).
					Joins(
						jn.Join[int](&db.Person.Id, &db.PersonJobTitle.IdPerson),
						jn.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(wh.Equals(&db.JobTitle.Id, jobs[0].Id)).OrderByAsc(&db.Person.Id).Rows() {

					if err != nil {
						t.Fatalf("Expected a select, got error: %v", err)
					}
					pj = append(pj, struct {
						JobTitle  string
						Person    string
						CreatedAt time.Time
					}{
						JobTitle:  query.SafeGet(row.JobTitle),
						Person:    query.SafeGet(row.Person),
						CreatedAt: query.SafeGet(row.CreatedAt),
					})
				}

				if len(pj) != 3 {
					t.Errorf("Expected %v, got : %v", 3, len(pj))
				}

				tm := time.Time{}
				if pj[len(pj)-1].CreatedAt.Unix() == tm.Unix() {
					t.Errorf("Expected value, got %v", pj[len(pj)-1].CreatedAt.Unix())
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
				err = query.Update(db.DB, db.DB).Includes(db.DB).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(goe.DB{})
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected a goe.ErrInvalidArg, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Invalid_Includes",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = query.Update(db.DB, db.DB).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(goe.DB{})
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
				err = query.Update(db.DB, db.Animal).Where(wh.Equals(&a.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrInvalidWhere) {
					t.Errorf("Expected a goe.ErrInvalidWhere, got error: %v", err)
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
				err = query.UpdateContext(ctx, db.DB, db.Animal).Includes(&db.Animal.Name).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(a)
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
				err = query.UpdateContext(ctx, db.DB, db.Animal).Includes(&db.Animal.Name).Where(wh.Equals(&db.Animal.Id, a.Id)).Value(a)
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
