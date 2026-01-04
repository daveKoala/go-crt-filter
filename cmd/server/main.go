package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/daveKoala/go-crt-filter/internal/api"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/status", api.StatusHandler).Methods("GET")
	router.HandleFunc("/scan", api.StatusHandler).Methods("POST")
	router.HandleFunc("/test", api.StatusHandler).Methods("POST")

	log.Println("🚀 Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
