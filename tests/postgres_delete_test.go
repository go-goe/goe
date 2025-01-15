package tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/olauro/goe"
	"github.com/olauro/goe/query"
	"github.com/olauro/goe/wh"
)

func TestPostgresDelete(t *testing.T) {
	db, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}
	if db.ConnPool.Stats().InUse != 0 {
		t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
	}
	err = query.Delete(db.DB, db.AnimalFood).Where()
	if err != nil {
		t.Fatalf("Expected delete AnimalFood, got error: %v", err)
	}
	err = query.Delete(db.DB, db.Flag).Where()
	if err != nil {
		t.Fatalf("Expected delete flags, got error: %v", err)
	}
	err = query.Delete(db.DB, db.Animal).Where()
	if err != nil {
		t.Fatalf("Expected delete animals, got error: %v", err)
	}
	err = query.Delete(db.DB, db.Food).Where()
	if err != nil {
		t.Fatalf("Expected delete foods, got error: %v", err)
	}
	err = query.Delete(db.DB, db.Habitat).Where()
	if err != nil {
		t.Fatalf("Expected delete habitats, got error: %v", err)
	}
	err = query.Delete(db.DB, db.Info).Where()
	if err != nil {
		t.Fatalf("Expected delete infos, got error: %v", err)
	}
	err = query.Delete(db.DB, db.Status).Where()
	if err != nil {
		t.Fatalf("Expected delete status, got error: %v", err)
	}
	err = query.Delete(db.DB, db.UserRole).Where()
	if err != nil {
		t.Fatalf("Expected delete user roles, got error: %v", err)
	}
	err = query.Delete(db.DB, db.User).Where()
	if err != nil {
		t.Fatalf("Expected delete users, got error: %v", err)
	}
	err = query.Delete(db.DB, db.Role).Where()
	if err != nil {
		t.Fatalf("Expected delete roles, got error: %v", err)
	}

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Delete_One_Record",
			testCase: func(t *testing.T) {
				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				a := Animal{Name: "Dog"}
				err = query.Insert(db.DB, db.Animal).One(&a)
				if err != nil {
					t.Fatalf("Expected a insert animal, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				var as *Animal
				as, err = query.Find(db.DB, db.Animal, Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				err = query.Remove(db.DB, db.Animal, Animal{Id: as.Id})
				if err != nil {
					t.Errorf("Expected a delete animal, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				_, err = query.Find(db.DB, db.Animal, Animal{Id: as.Id})
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}
			},
		},
		{
			desc: "Delete_All_Records",
			testCase: func(t *testing.T) {
				animals := []Animal{
					{Name: "Cat"},
					{Name: "Forest Cat"},
					{Name: "Catt"},
				}
				err = query.Insert(db.DB, db.Animal).All(animals)
				if err != nil {
					t.Fatalf("Expected a insert, got error: %v", err)
				}

				animals = nil
				animals, err = query.Select(db.DB, db.Animal).From(db.Animal).Where(wh.Like(&db.Animal.Name, "%Cat%")).RowsAsSlice()
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if len(animals) != 3 {
					t.Errorf("Expected 3, got %v", len(animals))
				}

				err = query.Delete(db.DB, db.Animal).Where(wh.Like(&db.Animal.Name, "%Cat%"))
				if err != nil {
					t.Fatalf("Expected a delete, got error: %v", err)
				}

				animals = nil
				animals, err = query.Select(db.DB, db.Animal).From(db.Animal).Where(wh.Like(&db.Animal.Name, "%Cat%")).RowsAsSlice()
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if len(animals) != 0 {
					t.Errorf(`Expected to delete all "Cat" animals, got: %v`, len(animals))
				}
			},
		},
		{
			desc: "Delete_Invalid_Arg",
			testCase: func(t *testing.T) {
				err = query.Delete(db.DB, db.DB).Where(wh.Equals(&db.Animal.Id, 1))
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected a goe.ErrInvalidArg, got error: %v", err)
				}
			},
		},
		{
			desc: "Delete_Invalid_Where",
			testCase: func(t *testing.T) {
				b := 2
				err = query.Delete(db.DB, db.Animal).Where(wh.Equals(&b, b))
				if !errors.Is(err, goe.ErrInvalidWhere) {
					t.Errorf("Expected a goe.ErrInvalidWhere, got error: %v", err)
				}
			},
		},
		{
			desc: "Delete_Context_Cancel",
			testCase: func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				err = query.DeleteContext(ctx, db.DB, db.Animal).Where()
				if !errors.Is(err, context.Canceled) {
					t.Errorf("Expected a context.Canceled, got error: %v", err)
				}
			},
		},
		{
			desc: "Delete_Context_Timeout",
			testCase: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond*1)
				defer cancel()
				err = query.DeleteContext(ctx, db.DB, db.Animal).Where()
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
