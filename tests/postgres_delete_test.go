package tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/olauro/goe"
	"github.com/olauro/goe/query"
)

func TestPostgresDelete(t *testing.T) {
	db, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}

	goeDb, err := goe.GetGoeDatabase(db)
	if err != nil {
		t.Fatalf("Expected goe database, got error: %v", err)
	}

	if goeDb.ConnPool.Stats().InUse != 0 {
		t.Errorf("Expected closed connection, got: %v", goeDb.ConnPool.Stats().InUse)
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

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Delete_One_Record",
			testCase: func(t *testing.T) {
				if goeDb.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", goeDb.ConnPool.Stats().InUse)
				}

				a := Animal{Name: "Dog"}
				err = goe.Insert(db.Animal).One(&a)
				if err != nil {
					t.Fatalf("Expected a insert animal, got error: %v", err)
				}

				if goeDb.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", goeDb.ConnPool.Stats().InUse)
				}

				var as *Animal
				as, err = goe.Find(db.Animal, Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if goeDb.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", goeDb.ConnPool.Stats().InUse)
				}

				err = goe.Remove(db.Animal, Animal{Id: as.Id})
				if err != nil {
					t.Errorf("Expected a delete animal, got error: %v", err)
				}

				if goeDb.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", goeDb.ConnPool.Stats().InUse)
				}

				_, err = goe.Find(db.Animal, Animal{Id: as.Id})
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
				err = goe.Insert(db.Animal).All(animals)
				if err != nil {
					t.Fatalf("Expected a insert, got error: %v", err)
				}

				animals = nil
				animals, err = goe.Select(db.Animal).From(db.Animal).Where(query.Like(&db.Animal.Name, "%Cat%")).RowsAsSlice()
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if len(animals) != 3 {
					t.Errorf("Expected 3, got %v", len(animals))
				}

				err = goe.Delete(db.Animal).Where(query.Like(&db.Animal.Name, "%Cat%"))
				if err != nil {
					t.Fatalf("Expected a delete, got error: %v", err)
				}

				animals = nil
				animals, err = goe.Select(db.Animal).From(db.Animal).Where(query.Like(&db.Animal.Name, "%Cat%")).RowsAsSlice()
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
				err = goe.Delete(&struct{}{}).Where(query.Equals(&db.Animal.Id, 1))
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected a goe.ErrInvalidArg, got error: %v", err)
				}
			},
		},
		{
			desc: "Delete_Invalid_Where",
			testCase: func(t *testing.T) {
				b := 2
				err = goe.Delete(db.Animal).Where(query.Equals(&b, b))
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
				err = goe.DeleteContext(ctx, db.Animal).Where()
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
				err = goe.DeleteContext(ctx, db.Animal).Where()
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
