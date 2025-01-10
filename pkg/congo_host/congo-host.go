package congo_host

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ccutch/congo/pkg/congo"
	"golang.org/x/crypto/ssh"
)

//go:embed all:migrations
var migrations embed.FS

type CongoHost struct {
	DB  *congo.Database
	api Platform
}

func InitCongoHost(root string, api Platform) *CongoHost {
	host := CongoHost{congo.SetupDatabase(root, "host.db", migrations), api}
	if err := host.DB.MigrateUp(); err != nil {
		log.Fatal("Failed to setup host db:", err)
	}
	if api != nil {
		api.Init(&host)
	}
	return &host
}

func (host *CongoHost) WithApi(api Platform) {
	if host.api = api; api != nil {
		host.api.Init(host)
	}
}

func (host *CongoHost) GenerateSSHKey(name string) (string, string, error) {
	serverData := filepath.Join(host.DB.Root, "hosts", name)
	os.MkdirAll(serverData, 0700)

	publicKeyPath := fmt.Sprintf("%s/id_rsa.pub", serverData)
	privateKeyPath := fmt.Sprintf("%s/id_rsa", serverData)
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
