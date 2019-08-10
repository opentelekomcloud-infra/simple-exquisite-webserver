package cmd

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Route mapping structure
type Route struct {
	Methods     []string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

func jsonResponse(w http.ResponseWriter, data []Entity) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func ok(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

var routes = []Route{
	Route{[]string{"GET"}, "/", ok},
}

// Router create new router from list of pre-defined routes
func Router() *mux.Router {
	router := mux.NewRouter()
	for _, route := range routes {
		router.Methods(route.Methods...)
	}
	return router
}
