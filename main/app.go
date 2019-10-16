package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/twinj/uuid"
)

const routeUUID4 = "/entity/{id:[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}}"

//App struct
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func generateRandomInitData(db *sql.DB) {
	ents := CreateSomeEntities(10000, 20000)
	err := AddEntities(db, ents)
	if err != nil {
		log.Println("Can't fill database with initial data")
	}
}

//Initialize func: init server according configuration structure
func (a *App) Initialize(config *Configuration) {
	if !config.Debug {
		var PgDbURL = config.Postgres.DbURL
		dbURLSliced := strings.Split(PgDbURL, ":")
		host := dbURLSliced[0]
		port, err := strconv.Atoi(dbURLSliced[1])
		if err != nil {
			log.Fatal(err)
		}
		createErr := CreatePostgreDBIfNotExist(config.Postgres.Database, host, port, config.Postgres.Username, config.Postgres.Database)
		if createErr != nil {
			log.Fatalf("Error during db creation: %v", createErr)
		}
		connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, config.Postgres.Username, config.Postgres.Password, config.Postgres.Database)
		a.DB, err = sql.Open("postgres", connectionString)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		a.DB = nil
	}
	CreateTable(a.DB)
	go generateRandomInitData(a.DB)
	a.Router = mux.NewRouter()
	a.InitializeRoutes()
}

//Run server
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

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

	if count > 10 || count < 1 {
		count = 10
	}

	entities, err := getEntities(a.DB, count, "%")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
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
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer func() { _ = r.Body.Close() }()

	if err := e.createEntity(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
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
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer func() { _ = r.Body.Close() }()

	if err := data.updateEntity(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
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
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
