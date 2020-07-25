package models

// RegisterRequest is a request the client makes to the API to create
// a new user.
type RegisterRequest struct {
	PublicKey string `json:"public_key"`
}

// GetRequest is a request the client makes to the API to get the public
// key of a recipient.
type GetRequest struct {
	ID string `json:"id"`
}

// SetRequest is a request the client makes to the API to store data.
type SetRequest struct {
	ID   string   `json:"id"`
	Data []byte   `json:"data"`
	TTL  Duration `json:"ttl"`
}
