package digitalocean

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

func (s *Server) keys() (string, string) {
	dir := filepath.Join(s.client.host.DB.Root, "hosts", s.Name)
	pubKey := fmt.Sprintf("%s/id_rsa.pub", dir)
	priKey := fmt.Sprintf("%s/id_rsa", dir)
	return pubKey, priKey
}

func (s *Server) setupAccess() error {
	if _, _, err := s.client.host.GenerateSSHKey(s.Name); err != nil {
		return errors.Wrap(err, "failed to create access key")
	}
	pubKey, _ := s.keys()
	data, err := os.ReadFile(pubKey)
	if err != nil {
		return errors.Wrap(err, "failed to read public key")
	}
	s.sshKey, _, err = s.client.Keys.Create(context.TODO(), &godo.KeyCreateRequest{
		Name:      s.Name + "-admin-key",
		PublicKey: string(data),
	})
	return errors.Wrap(err, "failed to create access key")
}

func (s *Server) deleteRemoteKeys() error {
	if s.sshKey == nil {
		return nil
	}
	fmt.Printf("Deleting SSH key %s...\n", s.sshKey.Name)
	if _, err := s.client.Keys.DeleteByID(context.TODO(), s.sshKey.ID); err != nil {
		return fmt.Errorf("failed to delete SSH key: %w", err)
	}
	s.sshKey = nil
	return nil
}

func (s *Server) deleteLocalKeys() error {
	pubKey, priKey := s.keys()
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
