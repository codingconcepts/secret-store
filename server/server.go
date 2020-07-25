package server

import (
	"encoding/json"
	"log"
	"net/http"
	"secret-store/pkg/models"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Server holds the runtime configuration for the API.
type Server struct {
	client redis.Cmdable
}

// New returns a pointer to a new Server.
func New(client redis.Cmdable) *Server {
	return &Server{
		client: client,
	}
}

// Register registers a new user.
func (s *Server) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		id := uuid.New().String()
		if err := s.client.Set(r.Context(), "user:"+id, request.PublicKey, 0).Err(); err != nil {
			respond(http.StatusInternalServerError, w, models.Response{Data: err.Error()})
			return
		}

		respond(http.StatusOK, w, models.Response{Data: id})
	}
}

// GetPublicKey returns the public key of a user by a given id.
func (s *Server) GetPublicKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			http.Error(w, "missing id parameter", http.StatusUnprocessableEntity)
			return
		}

		data, err := s.client.Get(r.Context(), "user:"+id).Result()
		if err != nil {
			log.Printf("error getting public key: %v", err)
			http.Error(w, "error getting public key", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(models.Response{Data: data}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Set sends a secret to someone.
func (s *Server) Set() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models.SetRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		if err := s.client.Set(r.Context(), "data:"+request.ID, request.Data, request.TTL.Duration).Err(); err != nil {
			log.Printf("error setting secret: %v", err)
			http.Error(w, "error setting secret", http.StatusInternalServerError)
			return
		}
	}
}

// Get receives a secret from someone.
func (s *Server) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			http.Error(w, "missing id parameter", http.StatusUnprocessableEntity)
			return
		}

		data, err := s.client.Get(r.Context(), id).Result()
		if err != nil {
			log.Printf("error getting secret: %v", err)
			http.Error(w, "error getting secret", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(models.Response{Data: data}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// respond writes a JSON response to to the caller, by marshalling the val struct.
func respond(code int, w http.ResponseWriter, val interface{}) {
	if code == http.StatusNoContent || val == nil {
		w.WriteHeader(code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	// If an error occurs, log it, as it's likely we won't be able to respond
	// to the user.
	if err := json.NewEncoder(w).Encode(val); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}
