package server

import (
	"encoding/json"
	"net/http"
	"secret-store/pkg/requests"

	"github.com/go-redis/redis/v8"
)

// Server holds the runtime configuration for the API.
type Server struct {
	db *redis.Cmdable
}

func (s *Server) register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request requests.Register
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}
}

func (s *Server) set() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
