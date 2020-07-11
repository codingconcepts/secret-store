package main

import (
	"github.com/go-redis/redis/v8"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	db, err := bolt.Open("./db", 0666, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}

	s := &server{
		db: db,
	}

	router := mux.NewRouter()
	router.HandleFunc("/", s.set()).Methods(http.MethodPost)
	router.HandleFunc("/{id}", s.get()).Methods(http.MethodGet)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: time.Second * 1,
		ReadTimeout:       time.Second * 1,
		WriteTimeout:      time.Second * 1,
		IdleTimeout:       time.Second * 5,
	}

	log.Fatal(server.ListenAndServe())
}
