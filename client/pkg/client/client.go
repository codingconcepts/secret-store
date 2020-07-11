package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client executes HTTP requests against the API.
type Client struct {
	net *http.Client
}

// New returns a pointer to a new Client.
func New() *Client {
	return &Client{
		net: &http.Client{
			Timeout: time.Second * 5,
		},
	}
}

// Execute makes a call to the API, marshalling a request into JSON if provided,
// and unmarshalling into a response object if provided.
// It returns a response code and an error if one occurred.
func (c *Client) Execute(ctx context.Context, method string, address string, request interface{}, response interface{}) error {
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

	req, err := http.NewRequestWithContext(ctx, method, address, bodyReader)
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
