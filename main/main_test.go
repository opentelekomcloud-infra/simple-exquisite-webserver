package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/outcatcher/simple-exquisite-webserver/main"
	"github.com/twinj/uuid"
)

var a main.App

const tableCreationQuery = `
CREATE TABLE IF NOT EXISTS entities(
	Id TEXT NOT NULL PRIMARY KEY,
	Data TEXT
);
`

type entity struct {
	Id   string
	Data string
}

/**
 * Helper functions
 */
func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	if _, err := a.DB.Exec("DELETE FROM entities"); err != nil {
		log.Fatal(err)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
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
		if _, err := a.DB.Exec("INSERT INTO entities(Id, Data) VALUES($1, $2)", "Data "+strconv.Itoa(i), uuid.NewV4().String()); err != nil {
			log.Fatal(err)
		}
	}
}

/**
 * Test functions
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

	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func TestNonExistingEntity(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/entity/5520c7c6-bc87-4c13-a4bd-b682c5a88187", nil)
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
	var id = uuid.NewV4().String()
	payload := []byte((fmt.Sprintf(`{"Data": "test data", "Id": "%s"}`, id)))
	req, _ := http.NewRequest("POST", "/entity", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	var err = json.Unmarshal(response.Body.Bytes(), &m)
	checkErr(err)
	if m["Data"] != "test data" {
		t.Errorf("Expected product Data to be 'test data'. Got '%v'", m["Data"])
	}

	if m["Id"] != id {
		t.Errorf("Expected product Id to be '%s'. Got '%v'", id, m["Id"])
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

	var originalEntity = []*entity{}
	var err = json.Unmarshal(response.Body.Bytes(), &originalEntity)
	checkErr(err)

	payload := []byte((fmt.Sprintf(`{"Data": "test data - updated", "Id": "%s"}`, originalEntity[0].Id)))
	updateRoute := fmt.Sprintf("/entity/%s", originalEntity[0].Id)
	req, _ = http.NewRequest("PUT", updateRoute, bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	err = json.Unmarshal(response.Body.Bytes(), &m)
	checkErr(err)

	if m["Id"] != originalEntity[0].Id {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalEntity[0].Id, m["Id"])
	}

	if m["Data"] == originalEntity[0].Data {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalEntity[0].Data, m["Data"], m["Data"])
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

	deleteRoute := fmt.Sprintf("/entity/%s", originalEntity[0].Id)
	req, _ = http.NewRequest("DELETE", deleteRoute, nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	getRoute := fmt.Sprintf("/entity/%s", originalEntity[0].Id)
	req, _ = http.NewRequest("GET", getRoute, nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
