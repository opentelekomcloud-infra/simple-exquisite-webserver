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
CREATE TABLE IF NOT EXISTS entity(
	uuid TEXT NOT NULL PRIMARY KEY,
	data TEXT
);
`

type entity struct {
	Uuid string
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
	if _, err := a.DB.Exec("DELETE FROM entity"); err != nil {
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
		_, err := a.DB.Exec("INSERT INTO entity(data, uuid) VALUES($1, $2)",
			"Data "+strconv.Itoa(i),
			uuid.NewV4().String(),
		)
		if err != nil {
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
