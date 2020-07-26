package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"secret-store/client/pkg/client"

	"github.com/spf13/cobra"
)

var (
	c          *client.Client
	configPath string
	server     string
	bits       int
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "initialise client with a new key",
		Run: func(cmd *cobra.Command, args []string) {
			register(configPath)
		},
	}
	pushCmd := &cobra.Command{
		Use:   "push",
		Short: "push a message to someone",
		Long:  "Arg[0] = their id\nArg[1] = the message\n Arg[2] (optional) = a ttl duration",
		Run:   push,
		Args:  cobra.RangeArgs(2, 3),
	}
	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: "pull a message from someone",
		Long:  "pull a message from someone",
		Run:   pull,
		Args:  cobra.ExactArgs(1),
	}

	rootCmd := &cobra.Command{}
	rootCmd.PersistentFlags().StringVar(&configPath, "c", "secret-store.json", "config file path")
	rootCmd.PersistentFlags().StringVar(&server, "s", "https://sandbox-183716.nw.r.appspot.com", "address of the server")
	rootCmd.PersistentFlags().IntVar(&bits, "b", 3072, "RSA bit strength [1024, 2048, 3072, 4096]")

	rootCmd.AddCommand(initCmd, pushCmd, pullCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

type config struct {
	ID               string `json:"id"`
	PrivateKeyUnsafe string `json:"private_key"`
	PublicKey        string `json:"public_key"`
}

func register(configPath string) {
	c = client.New(server)

	file, err := os.Open(configPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("error opening config file: %v", err)
	}

	// Create crypto config.
	var config config
	if config.PrivateKeyUnsafe, config.PublicKey, err = generateRSA(); err != nil {
		log.Fatal(err)
	}

	// Store public key and get ID.
	if config.ID, err = c.Register(config.PublicKey); err != nil {
		log.Fatalf("error registering user: %v", err)
	}

	// Write config.
	if file, err = os.Create(configPath); err != nil {
		log.Fatalf("error storing config: %v", err)
	}
	if err = json.NewEncoder(file).Encode(config); err != nil {
		log.Fatalf("error writing config: %v", err)
	}

	log.Printf("id = %s", config.ID)
}

func push(cmd *cobra.Command, args []string) {
	c = client.New(server)
	id := args[0]
	data := args[1]

	var ttl time.Duration
	if len(args) == 3 {
		t, err := time.ParseDuration(args[2])
		if err != nil {
			log.Fatalf("error parsing ttl for message: %v", err)
		}
		ttl = t
	}

	// Fetch the recipient's public key.
	pubPEM, err := c.GetPublicKeyPEM(id)
	if err != nil {
		log.Fatalf("error getting reciplient's public key: %v", err)
	}

	pub, err := pemToPublicKey([]byte(pubPEM))
	if err != nil {
		log.Fatalf("error parsing recipient's public key: %v", err)
	}

	// Encrypt the message.
	ct, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, []byte(data), nil)
	if err != nil {
		log.Fatalf("error encrypting message: %v", err)
	}

	log.Println(ct)

	// Send the message.
	id, err = c.Send(id, ct, ttl)
	if err != nil {
		log.Fatalf("error sending message: %v", err)
	}

	log.Printf("%s", id)
}

func pull(cmd *cobra.Command, args []string) {
	c = client.New(server)

	id := args[0]

	// Load the private key.
	f, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("error opening config file: %v", err)
	}

	var conf config
	if err = json.NewDecoder(f).Decode(&conf); err != nil {
		log.Fatalf("error parsing config file: %v", err)
	}

	pri, err := pemToPrivateKey(conf.PrivateKeyUnsafe)
	if err != nil {
		log.Fatalf("error reading private key: %v", err)
	}

	// Get the encrypted message.
	ct, err := c.Get(id)
	if err != nil {
		log.Fatalf("error getting message: %v", err)
	}

	// Decrypt the message and print.
	pt, err := pri.Decrypt(nil, ct, &rsa.OAEPOptions{Hash: crypto.SHA256})
	if err != nil {
		log.Fatalf("error decrypting message: %v", err)
	}

	log.Println(string(pt))
}

func generateRSA() (string, string, error) {
	private, err := rsa.GenerateKey(rand.Reader, bits)
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
