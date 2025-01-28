package tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olauro/goe"
)

func TestInsert(t *testing.T) {
	db, err := SetupPostgres()
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
					Byte:    []byte{1, 2, 3},
				}
				err = goe.Insert(db.Flag).One(&f)
				if err != nil {
					t.Fatalf("Expected a insert, got error: %v", err)
				}

				fs, _ := goe.Find(db.Flag, Flag{Id: f.Id})

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
				if fs.Uint32 != f.Uint32 {
					t.Errorf("Expected %v, got : %v", f.Uint32, fs.Uint32)
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
			desc: "Insert_Animal_Tx_Commit",
			testCase: func(t *testing.T) {
				a := &Animal{Name: "Cat"}

				var tx *goe.Tx
				// defult level of isolation is sql.LevelSerializable
				tx, err = goe.BeginTx(db)
				if err != nil {
					t.Fatalf("Expected a tx, got error: %v", err)
				}
				defer tx.Rollback()

				err = goe.Insert(db.Animal, tx).One(a)
				if err != nil {
					tx.Rollback()
					t.Fatalf("Expected a insert, got error: %v", err)
				}
				if a.Id == 0 {
					tx.Rollback()
					t.Fatalf("Expected a Id value, got : %v", a.Id)
				}

				// get record before commit or not using tx, will result in a goe.ErrNotFound
				_, err = goe.Find(db.Animal, Animal{Id: a.Id})
				if !errors.Is(err, goe.ErrNotFound) {
					tx.Rollback()
					t.Fatalf("Expected a Id value, got : %v", a.Id)
				}

				// get using same tx
				_, err = goe.Find(db.Animal, Animal{Id: a.Id}, tx)
				if err != nil {
					t.Fatalf("Expected Find, got : %v", err)
				}

				err = tx.Commit()
				if err != nil {
					t.Fatalf("Expected Commit Tx, got : %v", err)
				}

				_, err = goe.Find(db.Animal, Animal{Id: a.Id})
				if err != nil {
					t.Fatalf("Expected Find, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Animal_Tx_RollBack",
			testCase: func(t *testing.T) {
				a := &Animal{Name: "Cat"}

				var tx *goe.Tx
				// defult level of isolation is sql.LevelSerializable
				tx, err = goe.BeginTx(db)
				if err != nil {
					t.Fatalf("Expected a tx, got error: %v", err)
				}
				defer tx.Rollback()

				err = goe.Insert(db.Animal, tx).One(a)
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
				_, err = goe.Find(db.Animal, Animal{Id: a.Id})
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

				err = goe.Insert(db.PersonJobTitle).One(&PersonJobTitle{IdJobTitle: j.Id, IdPerson: p.Id, CreatedAt: time.Now()})
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
			desc: "Insert_Invalid_Value",
			testCase: func(t *testing.T) {
				err = goe.Insert(db.Animal).One(nil)
				if !errors.Is(err, goe.ErrInvalidInsertValue) {
					t.Errorf("Expected goe.ErrInvalidInsertValue, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Empty_Batch",
			testCase: func(t *testing.T) {
				animals := []Animal{}
				err = goe.Insert(db.Animal).All(animals)
				if !errors.Is(err, goe.ErrEmptyBatchValue) {
					t.Errorf("Expected goe.ErrInvalidInsertBatchValue, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Arg",
			testCase: func(t *testing.T) {
				err = goe.Insert(&struct{}{}).One(nil)
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected goe.ErrInvalidArg, got : %v", err)
				}

				err = goe.Insert[any](nil).One(nil)
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected goe.ErrInvalidArg, got : %v", err)
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
