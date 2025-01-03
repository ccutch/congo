package congo_host

import (
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/digitalocean/godo"
)

func (server *Server) startDroplet(size string) {
	if server.Error != nil {
		return
	}

	server.droplet, _, server.Error = server.platform.Droplets.Create(server.ctx, &godo.DropletCreateRequest{
		Name:   server.Name,
		Region: server.Region,
		Size:   size,
		Image: godo.DropletCreateImage{
			Slug: "docker-20-04",
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{Fingerprint: server.sshKey.Fingerprint},
		},
		Volumes: []godo.DropletCreateVolume{
			{ID: server.volume.ID},
		},
		Backups: true,
	})

	if server.Error != nil {
		return
	}

	for server.IP == "" {
		fmt.Printf("Server Info: %v\n", server)
		time.Sleep(10 * time.Second)
		if server.Refresh(); server.Error != nil {
			log.Fatalf("Error checking status: %s", server.Error)
		}
	}

	time.Sleep(30 * time.Second)
}

func (server *Server) deleteDroplet() error {
	if server.droplet == nil {
		return nil
	}

	fmt.Printf("Deleting droplet %s...\n", server.droplet.Name)
	_, err := server.platform.Droplets.Delete(server.ctx, server.droplet.ID)
	if err != nil {
		return fmt.Errorf("failed to delete droplet: %w", err)
	}

	server.droplet = nil
	return nil
}
