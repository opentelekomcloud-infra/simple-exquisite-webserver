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
	// ListenAndServer needs a port string and Handler which requires ServeHTTP(ResponseWriter, *Request) method
	// The mux.Router implements ServeHTTP(response, *request)
	log.Fatal(http.ListenAndServe(":6669", a.Router))
}

//InitializeRoutes method
func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/entitys", a.getEntitys).Methods("GET")
	a.Router.HandleFunc("/entitys", a.createEntity).Methods("POST")
	a.Router.HandleFunc("/entity/{id:[0-9]+}", a.getEntity).Methods("GET")
	a.Router.HandleFunc("/entity/{id:[0-9]+}", a.updateEntity).Methods("PUT")
	a.Router.HandleFunc("/entity/{id:[0-9]+}", a.deleteEntity).Methods("DELETE")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) getEntity(w http.ResponseWriter, r *http.Request) {
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

func (a *App) getEntitys(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.URL.Query().Get("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	entitys, err := getEntitys(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, entitys)
}

func (a *App) createEntity(w http.ResponseWriter, r *http.Request) {
	var e entity
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

func (a *App) updateEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	var e entity
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

func (a *App) deleteEntity(w http.ResponseWriter, r *http.Request) {
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
