package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"secret-store/pkg/models"
	"time"
)

// Client executes HTTP requests against the API.
type Client struct {
	Addr string
	net  *http.Client
}

// New returns a pointer to a new Client.
func New(addr string) *Client {
	return &Client{
		Addr: addr,
		net: &http.Client{
			Timeout: time.Second * 5,
		},
	}
}

// Register stores a user's public key and returns their server-generated ID.
func (c *Client) Register(pubPEM string) (string, error) {
	fullURL, err := joinURL(c.Addr, "users")
	if err != nil {
		return "", fmt.Errorf("error joining url: %w", err)
	}

	request := models.RegisterRequest{
		PublicKey: pubPEM,
	}

	// Make request to register user in the server and get the resulting ID.
	var response models.Response
	if err = c.Execute(http.MethodPost, fullURL, request, &response); err != nil {
		return "", fmt.Errorf("error registering client: %w", err)
	}
	return response.Data, err
}

// GetPublicKeyPEM returns the public key of a recipient with a given id.
func (c *Client) GetPublicKeyPEM(id string) (string, error) {
	fullURL, err := joinURL(c.Addr, "users/"+id)
	if err != nil {
		return "", fmt.Errorf("error joining url: %w", err)
	}

	var response models.Response
	if err = c.Execute(http.MethodGet, fullURL, nil, &response); err != nil {
		return "", fmt.Errorf("error getting public key: %w", err)
	}
	return response.Data, err
}

// Send sends an encrypted message to a user with a given id.
func (c *Client) Send(id string, data []byte, ttl time.Duration) error {
	fullURL, err := joinURL(c.Addr, "secrets")
	if err != nil {
		return fmt.Errorf("error joining url: %w", err)
	}

	// Make request to register user in the server and get the resulting ID.
	request := models.SetRequest{
		ID:   id,
		Data: data,
		TTL: models.Duration{
			Duration: ttl,
		},
	}

	if err = c.Execute(http.MethodPost, fullURL, request, nil); err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}
	return nil
}

// Execute makes a call to the API, marshalling a request into JSON if provided,
// and unmarshalling into a response object if provided.
// It returns a response code and an error if one occurred.
func (c *Client) Execute(method string, address string, request interface{}, response interface{}) error {
	// If this is a request with a body, marshal the request JSON and initialise
	// a reader, if not, the reader stays nil, which is expected.
	var bodyReader io.Reader
	if request != nil {
		body, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("error marshalling request: %w", err)
		}
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, address, bodyReader)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// All requests require the vnd.api+json content type header.
	req.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := c.net.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}

	// If we've been given nothing to bind to, or there's no response body from
	// the API, we're done.
	if response == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	return c.unmarshalBody(resp, response)
}

func (c *Client) unmarshalBody(r *http.Response, response interface{}) error {
	// We'll always try to read a response, so always try to close.
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(response); err != nil {
		return fmt.Errorf("error decoding response body: %v", err)
	}

	// It's safe to call Close multiple times, so in the case that the response body
	// is successfully unmarshalled, call it here.
	if err := r.Body.Close(); err != nil {
		return fmt.Errorf("error closing response body: %w", err)
	}

	return nil
}

func joinURL(base string, fragment string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, fragment)

	return u.String(), nil
}
