package congo_host

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/digitalocean/godo"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

type CongoHost struct {
	root     string
	platform *godo.Client

	*apiCreds
}

func InitCongoHost(root string, opts ...CongoHostOpt) *CongoHost {
	host := CongoHost{root: root}
	for _, opt := range opts {
		if err := opt(&host); err != nil {
			log.Fatal("Failed to setup CongoHost:", err)
		}
	}
	return &host
}

type CongoHostOpt func(*CongoHost) error

func WithApiToken(apiKey string) CongoHostOpt {
	if apiKey == "" {
		log.Fatal("Missing Digital Ocean API Key")
	}
	return func(host *CongoHost) error {
		host.WithApiToken(apiKey)
		return nil
	}
}

func (host *CongoHost) WithApiToken(token string) {
	if token == "" {
		host.platform = nil
	} else {
		host.platform = godo.NewClient(oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
		))
	}
}

func WithApiClient(clientID, secret, redirectURI string) CongoHostOpt {
	return func(host *CongoHost) error {
		host.WithApiClient(clientID, secret, redirectURI)
		return nil
	}
}

func (host *CongoHost) WithApiClient(clientID, secret, redirectURI string) {
	if clientID == "" || secret == "" || redirectURI == "" {
		host.apiCreds = nil
	} else {
		host.apiCreds = &apiCreds{clientID, secret, redirectURI}
	}
}

func (host *CongoHost) generateSSHKey(path string) (string, string, error) {
	serverData := filepath.Join(host.root, "hosts", path)
	os.MkdirAll(serverData, 0700)
	publicKeyPath := fmt.Sprintf("%s/id_rsa.pub", serverData)
	privateKeyPath := fmt.Sprintf("%s/id_rsa", serverData)
	// private key exists no need to proceed
	if _, err := os.Stat(privateKeyPath); err == nil {
		return publicKeyPath, privateKeyPath, nil
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return publicKeyPath, privateKeyPath, err
	}
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return publicKeyPath, privateKeyPath, err
	}
	defer privateKeyFile.Close()
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err = pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return publicKeyPath, privateKeyPath, err
	}
	if err = os.Chmod(privateKeyPath, 0600); err != nil {
		return publicKeyPath, privateKeyPath, err
	}
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return publicKeyPath, privateKeyPath, err
	}
	err = os.WriteFile(publicKeyPath, ssh.MarshalAuthorizedKey(publicKey), 0644)
	return publicKeyPath, privateKeyPath, err
}
