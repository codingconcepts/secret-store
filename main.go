package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.etcd.io/bbolt"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	db, err := bbolt.Open("secrets.db", 0666, nil)
	if err != nil {
		log.Fatalf("error opening boltdb: %v", err)
	}
	defer db.Close()

	s, err := server.New(db)
	if err != nil {
		log.Fatalf("error creating server: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/register", s.Register()).Methods(http.MethodPost)
	router.HandleFunc("/secrets", s.Set()).Methods(http.MethodPost)
	router.HandleFunc("/secrets/{id}", s.Get()).Methods(http.MethodGet)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: time.Second * 1,
		ReadTimeout:       time.Second * 1,
		WriteTimeout:      time.Second * 1,
		IdleTimeout:       time.Second * 5,
	}

	log.Println("server ready")
	log.Fatal(server.ListenAndServe())
}

func mustGetEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("missing env var: %q", key)
	}
	return value
}
