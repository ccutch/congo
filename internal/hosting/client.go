package hosting

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

type Client struct {
	platform *godo.Client
	dataPath string
}

func NewClient(path, apiKey string) *Client {
	token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})
	return &Client{
		platform: godo.NewClient(oauth2.NewClient(context.Background(), token)),
		dataPath: path,
	}
}

func (client *Client) GenerateSSHKey(name string) (string, string, error) {
	serverData := filepath.Join(client.dataPath, name)
	log.Println("server data", serverData)
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
