package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/opentelekomcloud-infra/simple-exquisite-webserver/main"
	"github.com/twinj/uuid"
)

var a main.App

type entity struct {
	Uuid string
	Data string
}

func clearTable() {
	main.FakeDataStorage = map[string]main.Entity{}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Fatalf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

func addEntities(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		err := main.FakeNew(&main.Entity{
			Uuid: uuid.NewV4().String(),
			Data: "Data " + strconv.Itoa(i),
		})

		if err != nil {
			log.Fatal(err)
		}
	}
}

/**
 * Test run setup/teardown
 */
func TestMain(m *testing.M) {
	a = main.App{}
	config, err := main.LoadConfiguration("")
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}
	config.Debug = true
	a.Initialize(config)

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func TestNonExistingEntity(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", fmt.Sprintf("/entity/%v", uuid.NewV4()), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	var err = json.Unmarshal(response.Body.Bytes(), &m)
	checkErr(err)

	if m["error"] != "entity not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'entity not found'. Got '%s'", m["error"])
	}
}

func TestCreateEntity(t *testing.T) {
	clearTable()
	payload := []byte(`{"data": "test data"}`)
	req, _ := http.NewRequest("POST", "/entity", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	var err = json.Unmarshal(response.Body.Bytes(), &m)
	checkErr(err)
	if m["data"] != "test data" {
		t.Errorf("Expected product Data to be 'test data'. Got '%v'", m["data"])
	}
}

func TestGetRoot(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetEntities(t *testing.T) {
	clearTable()
	addEntities(1)

	req, _ := http.NewRequest("GET", "/entities", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateEntity(t *testing.T) {
	clearTable()
	addEntities(1)

	req, _ := http.NewRequest("GET", "/entities", nil)
	response := executeRequest(req)

	var originalEntity []*entity
	var err = json.Unmarshal(response.Body.Bytes(), &originalEntity)
	checkErr(err)

	single := *originalEntity[0]
	payload := []byte(`{"data": "test data - updated"}`)
	updateRoute := fmt.Sprintf("/entity/%s", single.Uuid)
	req, _ = http.NewRequest("PUT", updateRoute, bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	err = json.Unmarshal(response.Body.Bytes(), &m)
	checkErr(err)

	if m["uuid"] != single.Uuid {
		t.Errorf("Expected the id to remain the same (%v). Got %v", single.Uuid, m["uuid"])
	}

	if m["data"] == single.Data {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", single.Data, m["data"], m["data"])
	}

}

func TestDeleteEntity(t *testing.T) {
	clearTable()
	addEntities(1)

	req, _ := http.NewRequest("GET", "/entities", nil)
	response := executeRequest(req)

	var originalEntity []*entity
	var err = json.Unmarshal(response.Body.Bytes(), &originalEntity)
	checkErr(err)

	checkResponseCode(t, http.StatusOK, response.Code)

	deleteRoute := fmt.Sprintf("/entity/%s", originalEntity[0].Uuid)
	req, _ = http.NewRequest("DELETE", deleteRoute, nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	getRoute := fmt.Sprintf("/entity/%s", originalEntity[0].Uuid)
	req, _ = http.NewRequest("GET", getRoute, nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestBulkDataGeneration(t *testing.T) {
	count := 10000
	size := 13
	entities := main.GenerateSomeEntities(count, size)
	rReal := len(entities)
	if rReal != count {
		t.Errorf("%d entities instead of %d", count, rReal)
	}
	for i := 0; i < count; i++ {
		data := entities[i].Data
		if len(data) != size {
			t.Errorf("One of entities size is not %d: %v", size, data)
		}
	}
}

func addSomeEntity(data ...string) error {
	var dataS string
	if len(data) > 0 {
		dataS = data[0]
	} else {
		dataS = main.RandomString(15, "MYDATA", okSet)
	}

	return main.FakeNew(&main.Entity{
		Uuid: uuid.NewV4().String(),
		Data: dataS,
	})
}

var okSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func TestGetEntitiesCount(t *testing.T) {
	max := 10
	addEntities(max + 3)
	randCount := rand.Intn(max) + 1
	r, _ := http.NewRequest("GET", fmt.Sprintf("/entities?count=%d", randCount), nil)
	response := executeRequest(r)
	var entities []*entity
	_ = json.Unmarshal(response.Body.Bytes(), &entities)
	if len(entities) != randCount {
		t.Error("Count is not limiting GetEntities")
	}
}

func TestGetEntitiesFilter(t *testing.T) {
	max := 10
	prefix := main.RandomString(3, "", okSet)
	for i := 0; i < max; i++ {
		data := main.RandomString(15, prefix, okSet)
		_ = addSomeEntity(data)
	}
	addEntities(max)
	r, _ := http.NewRequest("GET", fmt.Sprintf("/entities?filter=%s*", prefix), nil)
	response := executeRequest(r)
	var entities []*entity
	bts := response.Body.Bytes()
	_ = json.Unmarshal(bts, &entities)
	if len(entities) != max {
		t.Error("Filter is not limiting GetEntities")
	}
	for _, ent := range entities {
		if !strings.HasPrefix(ent.Data, prefix) {
			t.Errorf("Filter is not working: %s doesn't start with %s", ent.Data, prefix)
		}
	}
}
