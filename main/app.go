package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

//App struct
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

//LogWriter wrapper for error handling
type LogWriter struct {
	http.ResponseWriter
}

//Write func for LogWriter wrapper
func (w LogWriter) Write(p []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(p)
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
	return
}

//Initialize method
func (a *App) Initialize(user, password, dbname string) {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)
	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()
	a.InitializeRoutes()
}

//Run method
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

//InitializeRoutes method
func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/", a.Ok).Methods("GET")
	a.Router.HandleFunc("/entities", a.GetEntities).Methods("GET")
	a.Router.HandleFunc("/entity", a.CreateEntity).Methods("POST")
	a.Router.HandleFunc("/entity/{id:[0-9]+}", a.GetEntity).Methods("GET")
	a.Router.HandleFunc("/entity/{id:[0-9]+}", a.UpdateEntity).Methods("PUT")
	a.Router.HandleFunc("/entity/{id:[0-9]+}", a.DeleteEntity).Methods("DELETE")
}

func checkResponseOnError(w http.ResponseWriter, r *http.Request, e entity) {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	var writer = LogWriter{w}
	response, _ := json.Marshal(payload)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	writer.Write(response)
}

//Ok method
func (a *App) Ok(w http.ResponseWriter, r *http.Request) {
	var writer = LogWriter{w}
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

//GetEntity method
func (a *App) GetEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}
	e := entity{ID: id}

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

//CreateEntity method
func (a *App) CreateEntity(w http.ResponseWriter, r *http.Request) {
	var e entity
	checkResponseOnError(w, r, e)
	defer r.Body.Close()

	if err := e.createEntity(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, e)
}

//UpdateEntity method
func (a *App) UpdateEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	var e entity
	checkResponseOnError(w, r, e)
	defer r.Body.Close()
	e.ID = id

	if err := e.updateEntity(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, e)
}

//DeleteEntity method
func (a *App) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	e := entity{ID: id}
	if err := e.deleteEntity(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
