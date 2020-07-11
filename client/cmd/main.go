package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"

	"secret-store/client/pkg/client"
	"secret-store/pkg/models"
)

func main() {
	configPath := flag.String("p", "secret-store.json", "config file path")
	server := flag.String("s", "localhost:8080", "address of the server")
	flag.Parse()

	client := client.New(*server)

	// Load the config file from disk or register if it doesn't.
	config, err := initialise(client, *configPath)
	if err != nil {
		log.Fatalf("error initialising config: %v", err)
	}

	_ = config
}

type config struct {
	ID               string `json:"id"`
	PrivateKeyUnsafe string `json:"private_key"`
	PublicKey        string `json:"public_key"`
}

func initialise(c *client.Client, configPath string) (*config, error) {
	var config config

	if file, err := os.Open(configPath); err != nil {
		if err != os.ErrNotExist {
			return nil, fmt.Errorf("error opening config file: %w", err)
		}

		// Create config.
		if config.PrivateKeyUnsafe, config.PublicKey, err = generateRSA(); err != nil {
			return nil, err
		}
		config.ID = uuid.New().String()

		// Store config.
		request := models.RegisterRequest{
			PublicKey: config.PublicKey,
			ID:        config.ID,
		}

		if err = c.Execute(http.MethodPost, c.Addr, request, nil); err != nil {
			return nil, fmt.Errorf("error registering client: %w", err)
		}

	} else {
		if err = json.NewDecoder(file).Decode(&config); err != nil {
			return nil, fmt.Errorf("error decoding config file: %v", err)
		}
	}

	return &config, nil
}

func generateRSA() (string, string, error) {
	private, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", fmt.Errorf("error generating rsa keys: %w", err)
	}

	privatePEM := privateKeyToPEM(private)
	publicPEM, err := publicKeyToPEM(&private.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("error encoding public pem: %v", err)
	}

	return privatePEM, publicPEM, nil
}

func privateKeyToPEM(private *rsa.PrivateKey) string {
	privateBytes := x509.MarshalPKCS1PrivateKey(private)
	privatePEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateBytes,
		},
	)
	return string(privatePEM)
}

func pemToPrivateKey(pemBlock string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemBlock))
	if block == nil {
		return nil, fmt.Errorf("error parsing pem block")
	}

	private, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing pem")
	}

	return private, nil
}

func publicKeyToPEM(public *rsa.PublicKey) (string, error) {
	publicBytes, err := x509.MarshalPKIXPublicKey(public)
	if err != nil {
		return "", fmt.Errorf("error marshalling public key: %v", err)
	}
	publicPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicBytes,
		},
	)

	return string(publicPEM), nil
}

func pemToPublicKey(pemBlock []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBlock)
	if block == nil {
		return nil, fmt.Errorf("error parsing pem block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing pem")
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, fmt.Errorf("not an rsa key")
	}
}

func encrypt(publicPEM, message []byte) ([]byte, error) {
	public, err := pemToPublicKey(publicPEM)
	if err != nil {
		return nil, err
	}

	return rsa.EncryptOAEP(sha256.New(), rand.Reader, public, []byte(message), nil)
}
