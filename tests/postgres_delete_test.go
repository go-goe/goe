package tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/olauro/goe"
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

	err = db.Delete(db.AnimalFood).Where()
	if err != nil {
		t.Fatalf("Expected delete AnimalFood, got error: %v", err)
	}
	err = db.Delete(db.Flag).Where()
	if err != nil {
		t.Fatalf("Expected delete flags, got error: %v", err)
	}
	err = db.Delete(db.Animal).Where()
	if err != nil {
		t.Fatalf("Expected delete animals, got error: %v", err)
	}
	err = db.Delete(db.Food).Where()
	if err != nil {
		t.Fatalf("Expected delete foods, got error: %v", err)
	}
	err = db.Delete(db.Habitat).Where()
	if err != nil {
		t.Fatalf("Expected delete habitats, got error: %v", err)
	}
	err = db.Delete(db.Info).Where()
	if err != nil {
		t.Fatalf("Expected delete infos, got error: %v", err)
	}
	err = db.Delete(db.Status).Where()
	if err != nil {
		t.Fatalf("Expected delete status, got error: %v", err)
	}
	err = db.Delete(db.UserRole).Where()
	if err != nil {
		t.Fatalf("Expected delete user roles, got error: %v", err)
	}
	err = db.Delete(db.User).Where()
	if err != nil {
		t.Fatalf("Expected delete users, got error: %v", err)
	}
	err = db.Delete(db.Role).Where()
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
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				var as Animal
				err = db.Select(db.Animal).From(db.Animal).Where(wh.Equals(&db.Animal.Id, a.Id)).Scan(&as)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				err = db.Delete(db.Animal).Where(wh.Equals(&db.Animal.Id, as.Id))
				if err != nil {
					t.Errorf("Expected a delete animal, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				err = db.Select(db.Animal).From(db.Animal).Where(wh.Equals(&db.Animal.Id, a.Id)).Scan(&as)
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}

				var id int
				err = db.Select(&db.Animal.Id).From(db.Animal).Where(wh.Equals(&db.Animal.Id, a.Id)).Scan(&id)
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
				err = db.Insert(db.Animal).Value(&animals)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}

				animals = nil
				err = db.Select(db.Animal).From(db.Animal).Where(wh.Like(&db.Animal.Name, "%Cat%")).Scan(&animals)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}

				var a Animal
				err = db.Select(db.Animal).From(db.Animal).Where(wh.Like(&db.Animal.Name, "%Cat%")).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select one animal, got error: %v", err)
				}

				err = db.Delete(db.Animal).Where(wh.Like(&db.Animal.Name, "%Cat%"))
				if err != nil {
					t.Errorf("Expected a delete, got error: %v", err)
				}

				animals = nil
				err = db.Select(db.Animal).From(db.Animal).Where(wh.Like(&db.Animal.Name, "%Cat%")).Scan(&animals)
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}

				if len(animals) != 0 {
					t.Errorf(`Expected to delete all "Cat" animals, got: %v`, len(animals))
				}
			},
		},
		{
			desc: "Delete_Invalid_Arg",
			testCase: func(t *testing.T) {
				err = db.Delete(db.DB).Where(wh.Equals(&db.Animal.Id, 1))
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected a goe.ErrInvalidArg, got error: %v", err)
				}
			},
		},
		{
			desc: "Delete_Invalid_Where",
			testCase: func(t *testing.T) {
				b := 2
				err = db.Delete(db.Animal).Where(wh.Equals(&b, b))
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
				err = db.DeleteContext(ctx, db.Animal).Where()
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
				err = db.DeleteContext(ctx, db.Animal).Where()
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
