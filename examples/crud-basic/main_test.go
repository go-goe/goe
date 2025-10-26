package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-goe/examples/crud-basic/data"
	"github.com/go-goe/examples/crud-basic/handler"
	"github.com/go-goe/goe"
	"github.com/stretchr/testify/assert"
)

func TestApi(t *testing.T) {
	db, err := data.NewDatabase("crud-basic_test.db")
	assert.Nil(t, err)
	defer goe.Close(db)

	starter := frameworks[os.Getenv("PK")]
	assert.NotNil(t, starter)

	router, err := starter(db).Route()
	assert.Nil(t, err)
	defer os.Remove("crud-basic_test.db")

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Create_Find",
			testCase: func(t *testing.T) {

				buf := &bytes.Buffer{}
				json.NewEncoder(buf).Encode(data.Person{Name: "Lauro", Email: "email@teste.com"})

				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/persons", buf)
				router.ServeHTTP(w, r)

				assert.Equal(t, http.StatusCreated, w.Code)

				var res handler.Response[handler.ResponsePost[int]]
				json.NewDecoder(w.Body).Decode(&res)

				id := fmt.Sprint(res.Data.ID)
				w = httptest.NewRecorder()
				r = httptest.NewRequest("GET", "/persons/"+id, nil)
				router.ServeHTTP(w, r)

				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			desc: "Create_List",
			testCase: func(t *testing.T) {
				var buf *bytes.Buffer
				var w *httptest.ResponseRecorder
				var r *http.Request
				for range 10 {
					buf = &bytes.Buffer{}
					json.NewEncoder(buf).Encode(data.Person{Name: "Lauro", Email: "email@teste.com"})

					w = httptest.NewRecorder()
					r = httptest.NewRequest("POST", "/persons", buf)
					router.ServeHTTP(w, r)

					assert.Equal(t, http.StatusCreated, w.Code)
				}

				w = httptest.NewRecorder()
				r = httptest.NewRequest("GET", "/persons?page=1&size=5", nil)
				router.ServeHTTP(w, r)

				assert.Equal(t, http.StatusOK, w.Code)

				var res handler.Response[goe.Pagination[data.Person]]
				json.NewDecoder(w.Body).Decode(&res)

				assert.Len(t, res.Data.Values, 5)
			},
		},
		{
			desc: "Create_Remove",
			testCase: func(t *testing.T) {
				buf := &bytes.Buffer{}
				json.NewEncoder(buf).Encode(data.Person{Name: "Lauro", Email: "email@teste.com"})

				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/persons", buf)
				router.ServeHTTP(w, r)

				assert.Equal(t, http.StatusCreated, w.Code)

				var res handler.Response[handler.ResponsePost[int]]
				json.NewDecoder(w.Body).Decode(&res)

				id := fmt.Sprint(res.Data.ID)

				w = httptest.NewRecorder()
				r = httptest.NewRequest("DELETE", "/persons/"+id, nil)
				router.ServeHTTP(w, r)

				assert.Equal(t, http.StatusOK, w.Code)

				w = httptest.NewRecorder()
				r = httptest.NewRequest("GET", "/persons/"+id, nil)
				router.ServeHTTP(w, r)

				assert.Equal(t, http.StatusNotFound, w.Code)
			},
		},
		{
			desc: "Create_Save",
			testCase: func(t *testing.T) {
				buf := &bytes.Buffer{}
				json.NewEncoder(buf).Encode(data.Person{Name: "Lauro", Email: "email@teste.com"})

				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/persons", buf)
				router.ServeHTTP(w, r)

				assert.Equal(t, http.StatusCreated, w.Code)

				var res handler.Response[handler.ResponsePost[int]]
				json.NewDecoder(w.Body).Decode(&res)

				id := fmt.Sprint(res.Data.ID)

				w = httptest.NewRecorder()

				buf = &bytes.Buffer{}
				json.NewEncoder(buf).Encode(data.Person{Name: "Lauro Save"})

				r = httptest.NewRequest("PUT", "/persons/"+id, buf)
				router.ServeHTTP(w, r)

				assert.Equal(t, http.StatusOK, w.Code)

				w = httptest.NewRecorder()
				r = httptest.NewRequest("GET", "/persons/"+id, nil)
				router.ServeHTTP(w, r)

				assert.Equal(t, http.StatusOK, w.Code)

				var resFind handler.Response[data.Person]
				json.NewDecoder(w.Body).Decode(&resFind)

				assert.Equal(t, resFind.Data.Name, "Lauro Save")
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
