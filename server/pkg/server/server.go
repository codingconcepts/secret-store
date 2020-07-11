package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"secret-store/pkg/requests"
	"time"

	"github.com/go-redis/redis/v8"
)

// Server holds the runtime configuration for the API.
type Server struct {
	r redis.Cmdable
}

// New returns a pointer to a new Server.
func New(redisAddr, redisPass string) (*Server, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,
	})

	context, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if _, err := client.Ping(context).Result(); err != nil {
		return nil, fmt.Errorf("error pinging redis: %w", err)
	}

	return &Server{
		r: client,
	}, nil
}

// Register registers a new user.
func (s *Server) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request requests.Register
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}
}

// Set sends a secret to someone.
func (s *Server) Set() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// Get receives a secret from someone.
func (s *Server) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
