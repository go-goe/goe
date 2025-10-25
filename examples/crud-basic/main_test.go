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
	"github.com/go-goe/examples/crud-basic/framework/standard"
	"github.com/go-goe/examples/crud-basic/handler"
	"github.com/go-goe/goe"
	"github.com/stretchr/testify/assert"
)

func TestApi(t *testing.T) {
	db, err := data.NewDatabase("crud-basic_test.db")
	assert.Error(t, err)
	defer goe.Close(db)

	router, err := standard.Router(db)
	assert.Error(t, err)
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

				response := httptest.NewRecorder()
				request := httptest.NewRequest("POST", "/persons", buf)
				router.ServeHTTP(response, request)

				assert.Equal(t, http.StatusCreated, response.Code)

				var res handler.Response[handler.ResponsePost[int]]
				json.NewDecoder(response.Body).Decode(&res)

				id := fmt.Sprint(res.Data.ID)
				response = httptest.NewRecorder()
				request = httptest.NewRequest("GET", "/persons/"+id, nil)
				router.ServeHTTP(response, request)

				assert.Equal(t, http.StatusOK, response.Code)
			},
		},
		{
			desc: "Create_List",
			testCase: func(t *testing.T) {
				var buf *bytes.Buffer
				var response *httptest.ResponseRecorder
				var request *http.Request
				for range 10 {
					buf = &bytes.Buffer{}
					json.NewEncoder(buf).Encode(data.Person{Name: "Lauro", Email: "email@teste.com"})

					response = httptest.NewRecorder()
					request = httptest.NewRequest("POST", "/persons", buf)
					router.ServeHTTP(response, request)

					assert.Equal(t, http.StatusCreated, response.Code)
				}

				response = httptest.NewRecorder()
				request = httptest.NewRequest("GET", "/persons?page=1&size=5", nil)
				router.ServeHTTP(response, request)

				assert.Equal(t, http.StatusOK, response.Code)

				var res handler.Response[goe.Pagination[data.Person]]
				json.NewDecoder(response.Body).Decode(&res)

				assert.Len(t, res.Data.Values, 5)
			},
		},
		{
			desc: "Create_Remove",
			testCase: func(t *testing.T) {
				buf := &bytes.Buffer{}
				json.NewEncoder(buf).Encode(data.Person{Name: "Lauro", Email: "email@teste.com"})

				response := httptest.NewRecorder()
				request := httptest.NewRequest("POST", "/persons", buf)
				router.ServeHTTP(response, request)

				assert.Equal(t, http.StatusCreated, response.Code)

				var res handler.Response[handler.ResponsePost[int]]
				json.NewDecoder(response.Body).Decode(&res)

				id := fmt.Sprint(res.Data.ID)

				response = httptest.NewRecorder()
				request = httptest.NewRequest("DELETE", "/persons/"+id, nil)
				router.ServeHTTP(response, request)

				assert.Equal(t, http.StatusOK, response.Code)

				response = httptest.NewRecorder()
				request = httptest.NewRequest("GET", "/persons/"+id, nil)
				router.ServeHTTP(response, request)

				assert.Equal(t, http.StatusNotFound, response.Code)
			},
		},
		{
			desc: "Create_Save",
			testCase: func(t *testing.T) {
				buf := &bytes.Buffer{}
				json.NewEncoder(buf).Encode(data.Person{Name: "Lauro", Email: "email@teste.com"})

				response := httptest.NewRecorder()
				request := httptest.NewRequest("POST", "/persons", buf)
				router.ServeHTTP(response, request)

				assert.Equal(t, http.StatusCreated, response.Code)

				var res handler.Response[handler.ResponsePost[int]]
				json.NewDecoder(response.Body).Decode(&res)

				id := fmt.Sprint(res.Data.ID)

				response = httptest.NewRecorder()

				buf = &bytes.Buffer{}
				json.NewEncoder(buf).Encode(data.Person{Name: "Lauro Save"})

				request = httptest.NewRequest("PUT", "/persons/"+id, buf)
				router.ServeHTTP(response, request)

				assert.Equal(t, http.StatusOK, response.Code)

				response = httptest.NewRecorder()
				request = httptest.NewRequest("GET", "/persons/"+id, nil)
				router.ServeHTTP(response, request)

				assert.Equal(t, http.StatusOK, response.Code)

				var resFind handler.Response[data.Person]
				json.NewDecoder(response.Body).Decode(&resFind)

				assert.Equal(t, resFind.Data.Name, "Lauro Save")
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
