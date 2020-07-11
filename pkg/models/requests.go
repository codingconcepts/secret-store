package models

// RegisterRequest is a request to the client makes to the API to create
// a new user.
type RegisterRequest struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
}

// SetRequest is a request the client makes to the API to store data.
type SetRequest struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}
