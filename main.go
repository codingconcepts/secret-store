package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"secret-store/server"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     mustGetEnv("REDIS_ADDR"),
		Password: mustGetEnv("REDIS_PASS"),
		DB:       0,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("error pinging redis: %v", err)
	}

	s := server.New(client)

	router := mux.NewRouter()
	router.HandleFunc("/users", s.Register()).Methods(http.MethodPost)
	router.HandleFunc("/users/{id}", s.GetPublicKey()).Methods(http.MethodGet)
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
