package requests

// Register represents the public key and ID of a user.
type Register struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
}
