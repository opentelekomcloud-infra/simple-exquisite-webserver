package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/twinj/uuid"
)

//App struct
type App struct {
	Router           *mux.Router
	DB               *sql.DB
	DataGenerationWg sync.WaitGroup
}

func generateRandomInitData(db *sql.DB, config *Configuration, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	if config.Debug || config.Postgres == nil || config.Postgres.Initial == nil {
		log.Println("No initial data will be generated")
		return
	}

	initial := config.Postgres.Initial
	ents := GenerateSomeEntities(initial.Count, initial.Size)
	err := AddEntities(db, ents)
	if err != nil {
		log.Println("Can't fill database with initial data")
	}
}

//Initialize func: init server according configuration structure
func (a *App) Initialize(config *Configuration) {
	if !config.Debug {
		if config.Postgres == nil {
			log.Panic("No postgres configuration is given, but debug mode is disabled")
		}
		var PgDbURL = config.Postgres.DbURL
		dbURLSliced := strings.Split(PgDbURL, ":")
		host := dbURLSliced[0]
		port, err := strconv.Atoi(dbURLSliced[1])
		if err != nil {
			log.Fatal(err)
		}
		createErr := CreatePostgreDBIfNotExist(config.Postgres.Database, host, port, config.Postgres.Username, config.Postgres.Password)
		if createErr != nil {
			log.Fatalf("Error during db creation: %v", createErr)
		}
		connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, config.Postgres.Username, config.Postgres.Password, config.Postgres.Database)
		a.DB, err = sql.Open("postgres", connectionString)
		if err != nil {
			log.Fatal(err)
		}
		CreateTable(a.DB)
	} else {
		a.DB = nil
	}
	a.DataGenerationWg.Add(1)
	go generateRandomInitData(a.DB, config, &a.DataGenerationWg)
	a.Router = mux.NewRouter()
	a.InitializeRoutes()
}

//Run server
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

const routeUUID4 = "/entity/{id:[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}}"

//InitializeRoutes - init routes for api requests
func (a *App) InitializeRoutes() {
	a.Router.Use(addServerHeaderMiddle)
	a.Router.HandleFunc("/", a.Ok).Methods("GET")
	a.Router.HandleFunc("/entities", a.GetEntities).Methods("GET")
	a.Router.HandleFunc("/entity", a.CreateEntity).Methods("POST")
	a.Router.HandleFunc(routeUUID4, a.GetEntity).Methods("GET")
	a.Router.HandleFunc(routeUUID4, a.UpdateEntity).Methods("PUT")
	a.Router.HandleFunc(routeUUID4, a.DeleteEntity).Methods("DELETE")
}

func addServerHeaderMiddle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()
		w.Header().Set("Server", hostname)
		h.ServeHTTP(w, r)
	})
}

func logerr(_ int, err error) {
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, code int, err error) {
	log.Print(err)
	respondWithJSON(w, code, map[string]string{"error": err.Error()})
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
	logerr(w.Write(randomByteSlice(10, "OK", "0123456789abcdef")))
}

//GetEntity by Uuid
func (a *App) GetEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	e := Entity{Uuid: id}
	if err := e.getEntity(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, errors.New("entity not found"))
		default:
			respondWithError(w, http.StatusInternalServerError, err)
		}
		return
	}
	respondWithJSON(w, http.StatusOK, e)
}

const DefaultEntityListSize = 1000

//GetEntities Return list of entities, 1000 by default
func (a *App) GetEntities(w http.ResponseWriter, r *http.Request) {
	count := DefaultEntityListSize
	cStr := r.URL.Query().Get("count")
	if cStr != "" {
		count, _ = strconv.Atoi(cStr)
		if count < 1 {
			count = DefaultEntityListSize
		}
	}

	filter := r.URL.Query().Get("filter")
	entities, err := getEntities(a.DB, count, filter)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, entities)
}

//CreateEntity - with guid generator for Uuid's
func (a *App) CreateEntity(w http.ResponseWriter, r *http.Request) {
	var e Entity
	e.Uuid = uuid.NewV4().String()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("invalid request payload"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	if err := e.createEntity(a.DB); err != nil {
		log.Print(err)
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, e)
}

//UpdateEntity by Uuid
func (a *App) UpdateEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	data := Entity{vars["id"], ""}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("invalid request payload"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	if err := data.updateEntity(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, data)
}

//DeleteEntity by Uuid
func (a *App) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	e := Entity{Uuid: id}
	if err := e.deleteEntity(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
