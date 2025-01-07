package congo_host

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

func (server *Server) startDroplet(size string) {
	if server.Error != nil {
		return
	}

	ctx := context.Background()
	server.droplet, _, server.Error = server.host.platform.Droplets.Create(ctx, &godo.DropletCreateRequest{
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
		server.Refresh()
	}

	time.Sleep(30 * time.Second)
}

func (server *Server) deleteDroplet() error {
	if server.droplet == nil {
		return errors.New("server has no droplet")
	}

	ctx := context.Background()
	fmt.Printf("Deleting droplet %s...\n", server.droplet.Name)
	_, err := server.host.platform.Droplets.Delete(ctx, server.droplet.ID)
	if err != nil {
		return errors.Wrap(err, "failed to delete droplet")
	}

	server.droplet = nil
	return nil
}
