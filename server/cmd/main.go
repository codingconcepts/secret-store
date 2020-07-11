package main

import (
	"log"
	"net/http"
	"os"
	"secret-store/server/pkg/server"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	s, err := server.New(
		mustGetEnv("REDIS_ADDR"),
		mustGetEnv("REDIS_PASS"),
	)

	if err != nil {
		log.Fatalf("error creating server: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/register", s.Register()).Methods(http.MethodPost)
	router.HandleFunc("/", s.Set()).Methods(http.MethodPost)
	router.HandleFunc("/{id}", s.Get()).Methods(http.MethodGet)

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
