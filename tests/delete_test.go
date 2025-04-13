package tests_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-goe/goe"
	"github.com/go-goe/goe/query/where"
)

func TestDelete(t *testing.T) {
	db, err := Setup()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}

	if err != nil {
		t.Fatalf("Expected goe database, got error: %v", err)
	}

	if db.DB.Stats().InUse != 0 {
		t.Errorf("Expected closed connection, got: %v", db.DB.Stats().InUse)
	}
	err = goe.Delete(db.AnimalFood).Wheres()
	if err != nil {
		t.Fatalf("Expected delete AnimalFood, got error: %v", err)
	}
	err = goe.Delete(db.Flag).Wheres()
	if err != nil {
		t.Fatalf("Expected delete flags, got error: %v", err)
	}
	err = goe.Delete(db.Animal).Wheres()
	if err != nil {
		t.Fatalf("Expected delete animals, got error: %v", err)
	}
	err = goe.Delete(db.Food).Wheres()
	if err != nil {
		t.Fatalf("Expected delete foods, got error: %v", err)
	}
	err = goe.Delete(db.Habitat).Wheres()
	if err != nil {
		t.Fatalf("Expected delete habitats, got error: %v", err)
	}
	err = goe.Delete(db.Info).Wheres()
	if err != nil {
		t.Fatalf("Expected delete infos, got error: %v", err)
	}
	err = goe.Delete(db.Status).Wheres()
	if err != nil {
		t.Fatalf("Expected delete status, got error: %v", err)
	}
	err = goe.Delete(db.UserRole).Wheres()
	if err != nil {
		t.Fatalf("Expected delete user roles, got error: %v", err)
	}
	err = goe.Delete(db.User).Wheres()
	if err != nil {
		t.Fatalf("Expected delete users, got error: %v", err)
	}
	err = goe.Delete(db.Role).Wheres()
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
				if db.DB.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.DB.Stats().InUse)
				}

				a := Animal{Name: "Dog"}
				err = goe.Insert(db.Animal).One(&a)
				if err != nil {
					t.Fatalf("Expected a insert animal, got error: %v", err)
				}

				if db.DB.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.DB.Stats().InUse)
				}

				var as *Animal
				as, err = goe.Find(db.Animal).ById(Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if db.DB.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.DB.Stats().InUse)
				}

				err = goe.Remove(db.Animal).ById(Animal{Id: as.Id})
				if err != nil {
					t.Errorf("Expected a delete animal, got error: %v", err)
				}

				if db.DB.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.DB.Stats().InUse)
				}

				_, err = goe.Find(db.Animal).ById(Animal{Id: as.Id})
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}
			},
		},
		{
			desc: "Delete_One_Record_Tx_Rollback",
			testCase: func(t *testing.T) {
				a := Animal{Name: "Dog"}
				err = goe.Insert(db.Animal).One(&a)
				if err != nil {
					t.Fatalf("Expected a insert animal, got error: %v", err)
				}

				var as *Animal
				as, err = goe.Find(db.Animal).ById(Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				var tx goe.Transaction
				tx, err = db.NewTransaction()
				if err != nil {
					t.Fatalf("Expected tx, got error: %v", err)
				}
				defer tx.Rollback()

				err = goe.Remove(db.Animal).OnTransaction(tx).ById(Animal{Id: as.Id})
				if err != nil {
					t.Errorf("Expected a delete animal, got error: %v", err)
				}

				_, err = goe.Find(db.Animal).OnTransaction(tx).ById(Animal{Id: as.Id})
				if !errors.Is(err, goe.ErrNotFound) {
					tx.Rollback()
					t.Fatalf("Expected a select, got error: %v", err)
				}

				err = tx.Rollback()
				if err != nil {
					t.Fatalf("Expected Rollback, got error: %v", err)
				}

				_, err = goe.Find(db.Animal).ById(Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
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
				animals, err = goe.Select(db.Animal).From(db.Animal).Wheres(where.Like(&db.Animal.Name, "%Cat%")).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if len(animals) != 3 {
					t.Errorf("Expected 3, got %v", len(animals))
				}

				err = goe.Delete(db.Animal).Wheres(where.Like(&db.Animal.Name, "%Cat%"))
				if err != nil {
					t.Fatalf("Expected a delete, got error: %v", err)
				}

				animals = nil
				animals, err = goe.Select(db.Animal).From(db.Animal).Wheres(where.Like(&db.Animal.Name, "%Cat%")).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if len(animals) != 0 {
					t.Errorf(`Expected to delete all "Cat" animals, got: %v`, len(animals))
				}
			},
		},
		{
			desc: "Delete_All_Records_Tx_Commit",
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

				var tx goe.Transaction

				tx, err = db.NewTransaction()
				if err != nil {
					t.Fatalf("Expected tx, got error: %v", err)
				}
				defer tx.Rollback()

				err = goe.Delete(db.Animal).OnTransaction(tx).Wheres(where.Like(&db.Animal.Name, "%Cat%"))
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected a delete, got error: %v", err)
				}

				animals = nil
				animals, err = goe.Select(db.Animal).From(db.Animal).Wheres(where.Like(&db.Animal.Name, "%Cat%")).AsSlice()
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if len(animals) != 3 {
					t.Fatalf(`Expected 3 "Cat" animals, got: %v`, len(animals))
				}

				err = tx.Commit()
				if err != nil {
					t.Fatalf("Expected a Commit, got error: %v", err)
				}

				animals = nil
				animals, err = goe.Select(db.Animal).From(db.Animal).Wheres(where.Like(&db.Animal.Name, "%Cat%")).AsSlice()
				if err != nil {
					t.Fatalf("Expected a select, got error: %v", err)
				}

				if len(animals) != 0 {
					t.Fatalf(`Expected delete all "Cat" animals, got: %v`, len(animals))
				}
			},
		},
		{
			desc: "Delete_Race",
			testCase: func(t *testing.T) {
				var wg sync.WaitGroup
				for range 10 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						goe.Delete(db.PersonJobTitle).Wheres()
					}()
				}
				wg.Wait()
			},
		},
		{
			desc: "Delete_NotFound",
			testCase: func(t *testing.T) {
				err = goe.Remove(db.Animal).ById(Animal{Name: "Cat"})
				if !errors.Is(err, goe.ErrNotFound) {
					t.Fatalf("Expected goe.ErrNotFound, got %v", err)
				}
			},
		},
		{
			desc: "Remove_Custom_NotFound",
			testCase: func(t *testing.T) {
				customErr := errors.New("my custom error")
				err = goe.Remove(db.Animal).OnErrNotFound(customErr).ById(Animal{Name: "Create Cat"})
				if !errors.Is(err, customErr) {
					t.Fatalf("Expected customErr, got error: %v", err)
				}
			},
		},
		{
			desc: "Delete_Context_Cancel",
			testCase: func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				err = goe.DeleteContext(ctx, db.Animal).Wheres()
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
				err = goe.DeleteContext(ctx, db.Animal).Wheres()
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
