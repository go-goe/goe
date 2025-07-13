package tests_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-goe/goe"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestInsert(t *testing.T) {
	db, err := Setup()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Insert_Flag",
			testCase: func(t *testing.T) {
				f := Flag{
					Id:         uuid.New(),
					Name:       "Flag",
					Float32:    1.1,
					Float64:    2.2,
					Today:      time.Now(),
					Int:        -1,
					Int8:       -8,
					Int16:      -16,
					Int32:      -32,
					Int64:      -64,
					Uint:       1,
					Uint8:      8,
					Uint16:     16,
					Uint64:     64,
					Bool:       true,
					Byte:       []byte{1, 2, 3},
					NullId:     sql.Null[uuid.UUID]{V: uuid.New(), Valid: true},
					NullString: sql.NullString{String: "String Value", Valid: true},
					Price:      decimal.NewFromUint64(99),
				}
				err = goe.Insert(db.Flag).IgnoreFields(&db.Flag.Uint32).One(&f)
				if err != nil {
					t.Fatalf("Expected a insert, got error: %v", err)
				}

				fs, _ := goe.Find(db.Flag).ById(Flag{Id: f.Id})

				if fs.Id != f.Id {
					t.Errorf("Expected %v, got : %v", f.Id, fs.Id)
				}

				if fs.Name != f.Name {
					t.Errorf("Expected %v, got : %v", f.Name, fs.Name)
				}

				if fs.Float32 != f.Float32 {
					t.Errorf("Expected %v, got : %v", f.Float32, fs.Float32)
				}
				if fs.Float64 != f.Float64 {
					t.Errorf("Expected %v, got : %v", f.Float64, fs.Float64)
				}

				if fs.Today.Second() != f.Today.Second() {
					t.Errorf("Expected %v, got : %v", f.Today, fs.Today)
				}

				if fs.Int != f.Int {
					t.Errorf("Expected %v, got : %v", f.Int, fs.Int)
				}
				if fs.Int8 != f.Int8 {
					t.Errorf("Expected %v, got : %v", f.Int8, fs.Int8)
				}
				if fs.Int16 != f.Int16 {
					t.Errorf("Expected %v, got : %v", f.Int16, fs.Int16)
				}
				if fs.Int32 != f.Int32 {
					t.Errorf("Expected %v, got : %v", f.Int32, fs.Int32)
				}
				if fs.Int64 != f.Int64 {
					t.Errorf("Expected %v, got : %v", f.Int64, fs.Int64)
				}

				if fs.Uint != f.Uint {
					t.Errorf("Expected %v, got : %v", f.Uint, fs.Uint)
				}
				if fs.Uint8 != f.Uint8 {
					t.Errorf("Expected %v, got : %v", f.Uint8, fs.Uint8)
				}
				if fs.Uint16 != f.Uint16 {
					t.Errorf("Expected %v, got : %v", f.Uint16, fs.Uint16)
				}
				// check default value
				if fs.Uint32 != 32 {
					t.Errorf("Expected default %v, got : %v", 32, fs.Uint32)
				}
				if fs.Uint64 != f.Uint64 {
					t.Errorf("Expected %v, got : %v", f.Uint64, fs.Uint64)
				}

				if fs.Bool != f.Bool {
					t.Errorf("Expected %v, got : %v", f.Bool, fs.Bool)
				}

				if len(fs.Byte) != len(f.Byte) {
					t.Errorf("Expected %v, got : %v", len(f.Byte), len(fs.Byte))
				}

				if !fs.Price.Equal(f.Price) {
					t.Errorf("Expected %v, got : %v", f.Price, fs.Price)
				}

				if fs.NullId != f.NullId {
					t.Errorf("Expected %v, got : %v", f.NullId, fs.NullId)
				}

				if fs.NullString != f.NullString {
					t.Errorf("Expected %v, got : %v", f.NullString, fs.NullString)
				}
			},
		},
		{
			desc: "Insert_Animal",
			testCase: func(t *testing.T) {
				a := Animal{Name: "Cat"}
				err = goe.Insert(db.Animal).One(&a)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}
				if a.Id == 0 {
					t.Errorf("Expected a Id value, got : %v", a.Id)
				}
			},
		},
		{
			desc: "Insert_Race",
			testCase: func(t *testing.T) {
				var wg sync.WaitGroup
				for range 10 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						a := Animal{Name: "Cat"}
						goe.Insert(db.Animal).One(&a)
					}()
				}
				wg.Wait()
			},
		},
		{
			desc: "Insert_Animal_Tx_Commit",
			testCase: func(t *testing.T) {
				a := &Animal{Name: "Cat"}

				var tx goe.Transaction
				// defult level of isolation is sql.LevelSerializable
				tx, err = db.NewTransaction()
				if err != nil {
					t.Fatalf("Expected a tx, got error: %v", err)
				}
				defer tx.Rollback()

				err = goe.Insert(db.Animal).OnTransaction(tx).One(a)
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected a insert, got error: %v", err)
				}
				if a.Id == 0 {
					tx.Rollback()
					t.Fatalf("Expected a Id value, got : %v", a.Id)
				}

				// get record before commit or not using tx, will result in a goe.ErrNotFound
				_, err = goe.Find(db.Animal).ById(Animal{Id: a.Id})
				if !errors.Is(err, goe.ErrNotFound) {
					tx.Rollback()
					t.Fatalf("Expected a Id value, got : %v", a.Id)
				}

				// get using same tx
				_, err = goe.Find(db.Animal).OnTransaction(tx).ById(Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected Find, got : %v", err)
				}

				err = tx.Commit()
				if err != nil {
					t.Fatalf("Expected Commit Tx, got : %v", err)
				}

				_, err = goe.Find(db.Animal).ById(Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected Find, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Animal_Tx_RollBack",
			testCase: func(t *testing.T) {
				a := &Animal{Name: "Cat"}

				var tx goe.Transaction
				// defult level of isolation is sql.LevelSerializable
				tx, err = db.NewTransaction()
				if err != nil {
					t.Fatalf("Expected a tx, got error: %v", err)
				}
				defer tx.Rollback()

				err = goe.Insert(db.Animal).OnTransaction(tx).One(a)
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected a insert, got error: %v", err)
				}
				if a.Id == 0 {
					tx.Rollback()
					t.Fatalf("Expected a Id value, got : %v", a.Id)
				}

				err = tx.Rollback()
				if err != nil {
					t.Fatalf("Expected a tx Rollback, got error: %v", err)
				}

				// get record after rollback will result in a goe.ErrNotFound
				_, err = goe.Find(db.Animal).ById(Animal{Id: a.Id})
				if !errors.Is(err, goe.ErrNotFound) {
					t.Fatalf("Expected a Id value, got : %v", a.Id)
				}
			},
		},
		{
			desc: "Insert_Composed_Pk",
			testCase: func(t *testing.T) {
				p := Person{Name: "Jhon"}
				err = goe.Insert(db.Person).One(&p)
				if err != nil {
					t.Fatalf("Expected a insert person, got error: %v", err)
				}
				j := JobTitle{Name: "Developer"}
				err = goe.Insert(db.JobTitle).One(&j)
				if err != nil {
					t.Fatalf("Expected a insert job, got error: %v", err)
				}

				err = goe.Insert(db.PersonJobTitle).One(&PersonJobTitle{IdJobTitle: j.Id, PersonId: p.Id, CreatedAt: time.Now()})
				if err != nil {
					t.Errorf("Expected a insert PersonJobTitle, got error: %v", err)
				}
			},
		},
		{
			desc: "Insert_Batch_Animal",
			testCase: func(t *testing.T) {
				animals := []Animal{
					{Name: "Cat"},
					{Name: "Dog"},
					{Name: "Forest Cat"},
					{Name: "Bear"},
					{Name: "Lion"},
					{Name: "Puma"},
					{Name: "Snake"},
					{Name: "Whale"},
				}
				err = goe.Insert(db.Animal).All(animals)
				if err != nil {
					t.Fatalf("Expected insert animals, got error: %v", err)
				}
				for i := range animals {
					if animals[i].Id == 0 {
						t.Errorf("Expected a Id value, got : %v", animals[i].Id)
					}
				}
			},
		},
		{
			desc: "Insert_Context_Cancel",
			testCase: func(t *testing.T) {
				a := Animal{}
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				err = goe.InsertContext(ctx, db.Animal).One(&a)
				if !errors.Is(err, context.Canceled) {
					t.Errorf("Expected context.Canceled, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Context_Timeout",
			testCase: func(t *testing.T) {
				a := Animal{}
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				defer cancel()
				err = goe.InsertContext(ctx, db.Animal).One(&a)
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("Expected context.DeadlineExceeded, got : %v", err)
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
