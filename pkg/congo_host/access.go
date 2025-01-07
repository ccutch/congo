package congo_host

import (
	"context"
	"fmt"
	"os"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

func (server *Server) checkAccessKeys() {
	if server.Error != nil {
		return
	}
	pubKey, priKey := server.Keys()
	if _, server.Error = os.Stat(pubKey); server.Error != nil {
		return
	}
	_, server.Error = os.Stat(priKey)
}

func (server *Server) setupAccessKey() {
	if server.Error != nil {
		return
	}

	_, _, server.Error = server.host.generateSSHKey(server.Name)
	if server.Error != nil {
		return
	}

	pubKey, _ := server.Keys()
	data, err := os.ReadFile(pubKey)
	if err != nil {
		server.Error = errors.Wrap(err, "failed to read public key")
		return
	}

	ctx := context.Background()
	server.sshKey, _, server.Error = server.host.platform.Keys.Create(ctx, &godo.KeyCreateRequest{
		Name:      server.Name + "-admin-key",
		PublicKey: string(data),
	})
}

func (server *Server) deleteRemoteKeys() error {
	if server.sshKey == nil {
		return nil
	}

	ctx := context.Background()
	fmt.Printf("Deleting SSH key %s...\n", server.sshKey.Name)
	if _, err := server.host.platform.Keys.DeleteByID(ctx, server.sshKey.ID); err != nil {
		return fmt.Errorf("failed to delete SSH key: %w", err)
	}

	server.sshKey = nil
	return nil
}

func (server *Server) deleteLocalKeys() error {
	pubKey, priKey := server.Keys()
	if priKey != "" {
		if err := os.Remove(priKey); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete private key file: %w", err)
		}
	}

	if pubKey != "" {
		if err := os.Remove(pubKey); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete public key file: %w", err)
		}
	}

	return nil
}
