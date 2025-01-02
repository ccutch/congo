package congo_host

import (
	"fmt"
	"os"

	"github.com/digitalocean/godo"
)

func (server *Server) checkAccessKeys() {
	if server.Err != nil {
		return
	}
	if _, server.Err = os.Stat(server.pubKey); server.Err != nil {
		return
	}
	_, server.Err = os.Stat(server.priKey)
}

func (server *Server) setupAccessKey() {
	if server.Err != nil {
		return
	}
	server.pubKey, server.priKey, server.Err = server.GenerateSSHKey(server.Name)
	if server.Err != nil {
		return
	}
	var data []byte
	data, server.Err = os.ReadFile(server.pubKey)
	if server.Err != nil {
		return
	}
	server.sshKey, _, server.Err = server.CongoHost.platform.Keys.Create(server.ctx, &godo.KeyCreateRequest{
		Name:      server.Name + "-admin-key",
		PublicKey: string(data),
	})
}

func (server *Server) deleteSSHKey() error {
	if server.sshKey == nil {
		return nil
	}

	fmt.Printf("Deleting SSH key %s...\n", server.sshKey.Name)
	_, err := server.platform.Keys.DeleteByID(server.ctx, server.sshKey.ID)
	if err != nil {
		return fmt.Errorf("failed to delete SSH key: %w", err)
	}

	server.sshKey = nil
	return nil
}

func (server *Server) deleteLocalKeys() error {

	if server.priKey != "" {
		if err := os.Remove(server.priKey); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete private key file: %w", err)
		}
	}

	if server.pubKey != "" {
		if err := os.Remove(server.pubKey); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete public key file: %w", err)
		}
	}

	return nil
}
