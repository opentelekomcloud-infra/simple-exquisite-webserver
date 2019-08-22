package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/twinj/uuid"
)

//App struct
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

//Initialize func: init server according configuration structure
func (a *App) Initialize(config Configuration) {
	var err error
	if !config.Debug {
		var PgDbURL = config.PgDbURL
		dbURLSliced := strings.Split(PgDbURL, ":")
		port, err := strconv.Atoi(dbURLSliced[1])
		if err != nil {
			log.Fatal(err)
		}
		connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dbURLSliced[0], port, config.PgUsername, config.PgPassword, config.PgDatabase)
		a.DB, err = sql.Open("postgres", connectionString)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		a.DB, err = sql.Open("sqlite3", "entities.db")
		if err != nil {
			log.Fatal(err)
		}
		CreateTable(a.DB)
	}
	a.Router = mux.NewRouter()
	a.InitializeRoutes()
}

//Run server
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

//InitializeRoutes - init routes for api requests
func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/", a.Ok).Methods("GET")
	a.Router.HandleFunc("/entities", a.GetEntities).Methods("GET")
	a.Router.HandleFunc("/entity", a.CreateEntity).Methods("POST")
	a.Router.HandleFunc("/entity/{id:[a-z0-9_-]*}", a.GetEntity).Methods("GET")
	a.Router.HandleFunc("/entity/{id:[a-z0-9_-]*}", a.UpdateEntity).Methods("PUT")
	a.Router.HandleFunc("/entity/{id:[a-z0-9_-]*}", a.DeleteEntity).Methods("DELETE")
}

func logerr(n int, err error) {
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	logerr(w.Write(response))
}

//Ok answer for root calls
func (a *App) Ok(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	logerr(w.Write([]byte("OK")))
}

//GetEntity by Id
func (a *App) GetEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	e := entity{ID: id}
	fmt.Printf("App.go GetEntity ID: %v \n", e.ID)
	if err := e.getEntity(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "entity not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, e)
}

//GetEntities method
func (a *App) GetEntities(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.URL.Query().Get("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	entities, err := getEntities(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, entities)
}

//CreateEntity - with guid generator for Id's
func (a *App) CreateEntity(w http.ResponseWriter, r *http.Request) {
	var e entity
	e.ID = uuid.NewV4().String()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := e.createEntity(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, e)
}

//UpdateEntity by Id
func (a *App) UpdateEntity(w http.ResponseWriter, r *http.Request) {
	var e entity
	vars := mux.Vars(r)
	id := vars["id"]

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	e.ID = id

	if err := e.updateEntity(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, e)
}

//DeleteEntity by Id
func (a *App) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	e := entity{ID: id}
	if err := e.deleteEntity(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
