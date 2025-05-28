package tests_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-goe/goe"
	"github.com/go-goe/goe/query"
	"github.com/go-goe/goe/query/join"
	"github.com/go-goe/goe/query/update"
	"github.com/go-goe/goe/query/where"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestUpdate(t *testing.T) {
	db, err := Setup()
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
					Uint64:  64,
					Bool:    true,
					Byte:    []byte{1, 2, 3},
					Price:   decimal.NewFromUint64(99),
				}
				err = goe.Insert(db.Flag).One(&f)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}

				price, err := decimal.NewFromString("9.99")
				if err != nil {
					t.Errorf("Expected a price, got error: %v", err)
				}

				ff := Flag{
					Name:    "Flag_Test",
					Float32: 3.3,
					Float64: 4.4,
					Bool:    false,
					Byte:    []byte{1},
					Price:   price,
				}
				u := goe.Update(db.Flag).
					Sets(
						update.Set(&db.Flag.Name, ff.Name),
						update.Set(&db.Flag.Bool, ff.Bool))
				err = u.Sets(
					update.Set(&db.Flag.Float64, ff.Float64),
					update.Set(&db.Flag.Float32, ff.Float32),
					update.Set(&db.Flag.Price, ff.Price),
					update.Set(&db.Flag.Byte, ff.Byte)).
					Where(where.Equals(&db.Flag.Id, f.Id))
				if err != nil {
					t.Fatalf("Expected a update, got error: %v", err)
				}

				var fselect *Flag
				fselect, err = goe.Find(db.Flag).ById(Flag{Id: f.Id})
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
					t.Errorf("Expected a update on float64, got : %v", fselect.Float64)
				}
				if fselect.Bool != ff.Bool {
					t.Errorf("Expected a update on bool, got : %v", fselect.Bool)
				}
				if len(fselect.Byte) != len(ff.Byte) {
					t.Errorf("Expected a update on byte, got : %v", len(fselect.Byte))
				}
				if !fselect.Price.Equal(ff.Price) {
					t.Errorf("Expected a update on price, got : %v", fselect.Price)
				}
			},
		},
		{
			desc: "Save_Flag",
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
					Uint64:  64,
					Bool:    true,
					Byte:    []byte{1, 2, 3},
					Price:   decimal.NewFromUint64(99),
				}
				err = goe.Insert(db.Flag).One(&f)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}

				price, err := decimal.NewFromString("9.99")
				if err != nil {
					t.Errorf("Expected a price, got error: %v", err)
				}

				ff := Flag{
					Id:      f.Id,
					Name:    "Flag_Test",
					Float32: 3.3,
					Float64: 4.4,
					Byte:    []byte{1},
					Price:   price,
				}
				err = goe.Save(db.Flag).ByValue(ff)
				if err != nil {
					t.Fatalf("Expected a update, got error: %v", err)
				}

				var fselect *Flag
				fselect, err = goe.Find(db.Flag).ById(Flag{Id: f.Id})
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
					t.Errorf("Expected a update on float64, got : %v", fselect.Float64)
				}
				if len(fselect.Byte) != len(ff.Byte) {
					t.Errorf("Expected a update on byte, got : %v", len(fselect.Byte))
				}
				if !fselect.Price.Equal(ff.Price) {
					t.Errorf("Expected a update on price, got : %v", fselect.Price)
				}
			},
		},
		{
			desc: "Update_Race",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				err = goe.Insert(db.Animal).One(&a)
				if err != nil {
					t.Fatalf("Expected a insert animal, got error: %v", err)
				}
				var wg sync.WaitGroup
				for range 10 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						au := Animal{Id: a.Id}
						au.Name = "Update Cat"
						goe.Save(db.Animal).ByValue(au)
					}()
				}
				wg.Wait()
			},
		},
		{
			desc: "Update_Animal",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				err = goe.Insert(db.Animal).One(&a)
				if err != nil {
					t.Fatalf("Expected a insert animal, got error: %v", err)
				}

				w := Weather{
					Name: "Warm",
				}
				err = goe.Insert(db.Weather).One(&w)
				if err != nil {
					t.Fatalf("Expected a insert weather, got error: %v", err)
				}

				h := Habitat{
					Id:        uuid.New(),
					Name:      "City",
					IdWeather: w.Id,
				}
				err = goe.Insert(db.Habitat).One(&h)
				if err != nil {
					t.Fatalf("Expected a insert habitat, got error: %v", err)
				}

				a.IdHabitat = &h.Id
				a.Name = "Update Cat"
				err = goe.Save(db.Animal).ByValue(a)
				if err != nil {
					t.Fatalf("Expected a update, got error: %v", err)
				}

				var aselect *Animal
				aselect, err = goe.Find(db.Animal).ById(Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if aselect.IdHabitat == nil || *aselect.IdHabitat != h.Id {
					t.Errorf("Expected a update on id habitat, got : %v", aselect.IdHabitat)
				}
				if aselect.Name != "Update Cat" {
					t.Errorf("Expected a update on name, got : %v", aselect.Name)
				}

				aselect.IdHabitat = nil
				err = goe.Update(db.Animal).Sets(update.Set(&db.Animal.IdHabitat, aselect.IdHabitat)).
					Where(where.Equals(&db.Animal.Id, aselect.Id))
				if err != nil {
					t.Fatalf("Expected a update, got error: %v", err)
				}

				aselect, err = goe.Find(db.Animal).ById(Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if aselect.IdHabitat != nil {
					t.Errorf("Expected IdHabitat to be nil, got : %v", aselect.IdHabitat)
				}
			},
		},
		{
			desc: "Update_Animal_Tx_Commit",
			testCase: func(t *testing.T) {
				var tx goe.Transaction

				tx, err = db.NewTransaction()
				if err != nil {
					t.Fatalf("Expected tx, got error: %v", err)
				}

				a := Animal{
					Name: "Cat",
				}

				defer tx.Rollback()
				err = goe.Insert(db.Animal).OnTransaction(tx).One(&a)
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected a insert animal, got error: %v", err)
				}

				w := Weather{
					Name: "Warm",
				}
				err = goe.Insert(db.Weather).OnTransaction(tx).One(&w)
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected a insert weather, got error: %v", err)
				}

				h := Habitat{
					Id:        uuid.New(),
					Name:      "City",
					IdWeather: w.Id,
				}
				err = goe.Insert(db.Habitat).OnTransaction(tx).One(&h)
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected a insert habitat, got error: %v", err)
				}

				a.IdHabitat = &h.Id
				a.Name = "Update Cat"
				err = goe.Save(db.Animal).OnTransaction(tx).ByValue(a)
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected a update, got error: %v", err)
				}

				// get record before commit or not using tx, will result in a goe.ErrNotFound
				_, err = goe.Find(db.Animal).ById(Animal{Id: a.Id})
				if !errors.Is(err, goe.ErrNotFound) {
					tx.Rollback()
					t.Fatalf("Expected a goe.ErrNotFound, got error: %v", err)
				}

				err = tx.Commit()
				if err != nil {
					t.Fatalf("Expected Commit, got error: %v", err)
				}

				var aselect *Animal
				aselect, err = goe.Find(db.Animal).ById(Animal{Id: a.Id})

				if aselect.IdHabitat == nil || *aselect.IdHabitat != h.Id {
					t.Errorf("Expected a update on id habitat, got : %v", aselect.IdHabitat)
				}
				if aselect.Name != "Update Cat" {
					t.Errorf("Expected a update on name, got : %v", aselect.Name)
				}
			},
		},
		{
			desc: "Update_PersonJobs_Tx_Rollback",
			testCase: func(t *testing.T) {
				var tx goe.Transaction

				tx, err = db.NewTransaction()
				if err != nil {
					t.Fatalf("Expected tx, got error: %v", err)
				}
				defer tx.Rollback()

				persons := []Person{
					{Name: "Jhon"},
					{Name: "Laura"},
					{Name: "Luana"},
				}
				err = goe.Insert(db.Person).OnTransaction(tx).All(persons)
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected insert persons, got error: %v", err)
				}

				jobs := []JobTitle{
					{Name: "Developer"},
					{Name: "Designer"},
				}
				err = goe.Insert(db.JobTitle).OnTransaction(tx).All(jobs)
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected insert jobs, got error: %v", err)
				}

				personJobs := []PersonJobTitle{
					{PersonId: persons[0].Id, IdJobTitle: jobs[0].Id, CreatedAt: time.Now()},
					{PersonId: persons[1].Id, IdJobTitle: jobs[0].Id, CreatedAt: time.Now()},
					{PersonId: persons[2].Id, IdJobTitle: jobs[1].Id, CreatedAt: time.Now()},
				}
				err = goe.Insert(db.PersonJobTitle).OnTransaction(tx).All(personJobs)
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected insert personJobs, got error: %v", err)
				}

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
				}).OnTransaction(tx).
					Joins(
						join.Join[int](&db.Person.Id, &db.PersonJobTitle.PersonId),
						join.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(where.Equals(&db.JobTitle.Id, jobs[0].Id)).Rows() {

					if err != nil {
						t.Fatalf("Expected a select, got error: %v", err)
					}
					pj = append(pj, struct {
						JobTitle string
						Person   string
					}{
						JobTitle: query.Get(row.JobTitle),
						Person:   query.Get(row.Person),
					})
				}

				if len(pj) != 2 {
					t.Errorf("Expected %v, got : %v", 2, len(pj))
				}
				err = goe.Update(db.PersonJobTitle).OnTransaction(tx).Sets(update.Set(&db.PersonJobTitle.IdJobTitle, jobs[0].Id)).
					Where(
						where.And(
							where.Equals(&db.PersonJobTitle.PersonId, persons[2].Id),
							where.Equals(&db.PersonJobTitle.IdJobTitle, jobs[1].Id),
						),
					)

				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected a update, got error: %v", err)
				}

				pj = nil
				for row, err := range goe.Select(&struct {
					JobTitle *string
					Person   *string
				}{
					JobTitle: &db.JobTitle.Name,
					Person:   &db.Person.Name,
				}).OnTransaction(tx).
					Joins(
						join.Join[int](&db.Person.Id, &db.PersonJobTitle.PersonId),
						join.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(where.Equals(&db.JobTitle.Id, jobs[0].Id)).Rows() {

					if err != nil {
						t.Fatalf("Expected a select, got error: %v", err)
					}
					pj = append(pj, struct {
						JobTitle string
						Person   string
					}{
						JobTitle: query.Get(row.JobTitle),
						Person:   query.Get(row.Person),
					})
				}

				if len(pj) != 3 {
					t.Errorf("Expected %v, got : %v", 3, len(pj))
				}

				err = tx.Rollback()
				if err != nil {
					t.Fatalf("Expected Rollback, got error: %v", err)
				}

				pj = nil
				for row, err := range goe.Select(&struct {
					JobTitle *string
					Person   *string
				}{
					JobTitle: &db.JobTitle.Name,
					Person:   &db.Person.Name,
				}).
					Joins(
						join.Join[int](&db.Person.Id, &db.PersonJobTitle.PersonId),
						join.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(where.Equals(&db.JobTitle.Id, jobs[0].Id)).Rows() {

					if err != nil {
						t.Fatalf("Expected a select, got error: %v", err)
					}
					pj = append(pj, struct {
						JobTitle string
						Person   string
					}{
						JobTitle: query.Get(row.JobTitle),
						Person:   query.Get(row.Person),
					})
				}

				if len(pj) != 0 {
					t.Errorf("Expected %v, got : %v", 0, len(pj))
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
					{PersonId: persons[0].Id, IdJobTitle: jobs[0].Id, CreatedAt: time.Now()},
					{PersonId: persons[1].Id, IdJobTitle: jobs[0].Id, CreatedAt: time.Now()},
					{PersonId: persons[2].Id, IdJobTitle: jobs[1].Id, CreatedAt: time.Now()},
				}
				err = goe.Insert(db.PersonJobTitle).All(personJobs)
				if err != nil {
					t.Fatalf("Expected insert personJobs, got error: %v", err)
				}

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
				}).
					Joins(
						join.Join[int](&db.Person.Id, &db.PersonJobTitle.PersonId),
						join.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(where.Equals(&db.JobTitle.Id, jobs[0].Id)).Rows() {

					if err != nil {
						t.Fatalf("Expected a select, got error: %v", err)
					}
					pj = append(pj, struct {
						JobTitle string
						Person   string
					}{
						JobTitle: query.Get(row.JobTitle),
						Person:   query.Get(row.Person),
					})
				}

				if len(pj) != 2 {
					t.Errorf("Expected %v, got : %v", 2, len(pj))
				}

				err = goe.Update(db.PersonJobTitle).Sets(update.Set(&db.PersonJobTitle.IdJobTitle, jobs[0].Id)).Where(
					where.And(where.Equals(&db.PersonJobTitle.PersonId, persons[2].Id), where.Equals(&db.PersonJobTitle.IdJobTitle, jobs[1].Id)))
				if err != nil {
					t.Fatalf("Expected a update, got error: %v", err)
				}

				pj = nil
				for row, err := range goe.Select(&struct {
					JobTitle *string
					Person   *string
				}{
					JobTitle: &db.JobTitle.Name,
					Person:   &db.Person.Name,
				}).
					Joins(
						join.Join[int](&db.Person.Id, &db.PersonJobTitle.PersonId),
						join.Join[int](&db.JobTitle.Id, &db.PersonJobTitle.IdJobTitle),
					).
					Where(where.Equals(&db.JobTitle.Id, jobs[0].Id)).Rows() {

					if err != nil {
						t.Fatalf("Expected a select, got error: %v", err)
					}
					pj = append(pj, struct {
						JobTitle string
						Person   string
					}{
						JobTitle: query.Get(row.JobTitle),
						Person:   query.Get(row.Person),
					})
				}

				if len(pj) != 3 {
					t.Errorf("Expected %v, got : %v", 3, len(pj))
				}
			},
		},
		{
			desc: "Save_BadRequest",
			testCase: func(t *testing.T) {
				err = goe.Save(db.Animal).ByValue(Animal{Name: "Cat"})
				if !errors.Is(err, goe.ErrBadRequest) {
					t.Fatalf("Expected goe.ErrBadRequest, got %v", err)
				}
			},
		},
		{
			desc: "Save_Custom_BadRequest",
			testCase: func(t *testing.T) {
				customErr := errors.New("my custom error")
				err = goe.Save(db.Animal).OnErrBadRequest(customErr).ByValue(Animal{Name: "Create Cat"})
				if !errors.Is(err, customErr) {
					t.Fatalf("Expected customErr, got error: %v", err)
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
				err = goe.UpdateContext(ctx, db.Animal).Sets(update.Set(&db.Animal.Name, a.Name)).Where(where.Equals(&db.Animal.Id, a.Id))
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
				err = goe.UpdateContext(ctx, db.Animal).Sets(update.Set(&db.Animal.Name, a.Name)).Where(where.Equals(&db.Animal.Id, a.Id))
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
