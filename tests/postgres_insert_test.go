package tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olauro/goe"
	"github.com/olauro/goe/wh"
)

func TestPostgresInsert(t *testing.T) {
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
				}
				err = db.Insert(db.Flag).Value(&f)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}

				var fs Flag
				db.Select(db.Flag).From(db.Flag).Where(wh.Equals(&db.Flag.Id, f.Id)).Scan(&fs)

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
			},
		},
		{
			desc: "Insert_Animal",
			testCase: func(t *testing.T) {
				a := Animal{Name: "Cat"}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}
				if a.Id == 0 {
					t.Errorf("Expected a Id value, got : %v", a.Id)
				}
			},
		},
		{
			desc: "Insert_Composed_Pk",
			testCase: func(t *testing.T) {
				p := Person{Name: "Jhon"}
				err = db.Insert(db.Person).Value(&p)
				if err != nil {
					t.Errorf("Expected a insert person, got error: %v", err)
				}
				j := JobTitle{Name: "Developer"}
				err = db.Insert(db.JobTitle).Value(&j)
				if err != nil {
					t.Errorf("Expected a insert job, got error: %v", err)
				}

				err = db.Insert(db.PersonJobTitle).Value(&PersonJobTitle{IdJobTitle: j.Id, IdPerson: p.Id, CreatedAt: time.Now()})
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
				err = db.Insert(db.Animal).Value(&animals)
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
			desc: "Insert_Invalid_Pointer",
			testCase: func(t *testing.T) {
				a := Animal{Name: "Cat"}
				err = db.Insert(db.Animal).Value(a)
				if !errors.Is(err, goe.ErrInvalidInsertPointer) {
					t.Errorf("Expected goe.ErrInvalidInsertPointer, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Value",
			testCase: func(t *testing.T) {
				a := 2
				err = db.Insert(db.Animal).Value(&a)
				if !errors.Is(err, goe.ErrInvalidInsertValue) {
					t.Errorf("Expected goe.ErrInvalidInsertValue, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Batch_Value",
			testCase: func(t *testing.T) {
				animals := []int{
					1,
					2,
				}
				err = db.Insert(db.Animal).Value(&animals)
				if !errors.Is(err, goe.ErrInvalidInsertBatchValue) {
					t.Errorf("Expected goe.ErrInvalidInsertBatchValue, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Empty_Batch",
			testCase: func(t *testing.T) {
				animals := []Animal{}
				err = db.Insert(db.Animal).Value(&animals)
				if !errors.Is(err, goe.ErrEmptyBatchValue) {
					t.Errorf("Expected goe.ErrInvalidInsertBatchValue, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Arg",
			testCase: func(t *testing.T) {
				a := 2
				err = db.Insert(db.DB).Value(&a)
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected goe.ErrInvalidArg, got : %v", err)
				}

				err = db.Insert(nil).Value(&a)
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
				err = db.InsertContext(ctx, db.Animal).Value(&a)
				if !errors.Is(err, context.Canceled) {
					t.Errorf("Expected context.Canceled, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Context_Timeout",
			testCase: func(t *testing.T) {
				a := Animal{}
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond*1)
				defer cancel()
				err = db.InsertContext(ctx, db.Animal).Value(&a)
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
