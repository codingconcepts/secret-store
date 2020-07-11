package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.etcd.io/bbolt"
)

var (
	userBucket   = []byte("user")
	secretBucket = []byte("secret")
)

// Server holds the runtime configuration for the API.
type Server struct {
	db *bbolt.DB
}

// New returns a pointer to a new Server.
func New(db *bbolt.DB) (*Server, error) {
	return &Server{
		db: db,
	}, nil
}

type registerRequest struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
}

// Register registers a new user.
func (s *Server) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request registerRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		if err := s.set(userBucket, []byte(request.ID), []byte(request.PublicKey)); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}
}

type setRequest struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

// Set sends a secret to someone.
func (s *Server) Set() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request setRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		if err := s.set(secretBucket, []byte(request.ID), []byte(request.Data)); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
	}
}

type getResponse struct {
	Data string `json:"data"`
}

// Get receives a secret from someone.
func (s *Server) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			http.Error(w, "missing id parameter", http.StatusUnprocessableEntity)
			return
		}

		data := s.get(secretBucket, []byte(id))
		if data == nil {
			http.Error(w, "", http.StatusNoContent)
			return
		}

		if err := json.NewEncoder(w).Encode(getResponse{Data: string(data)}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) get(bucket, key []byte) []byte {
	var value []byte
	s.db.View(func(tx *bbolt.Tx) error {
		value = tx.Bucket(bucket).Get(key)
		return nil
	})

	return value
}

func (s *Server) set(bucket, key, value []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).Put(key, value)
	})
}