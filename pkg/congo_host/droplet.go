package congo_host

import (
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
		server.Refresh()
	}

	time.Sleep(30 * time.Second)
}

func (server *Server) deleteDroplet() error {
	if server.droplet == nil {
		return errors.New("server has no droplet")
	}

	fmt.Printf("Deleting droplet %s...\n", server.droplet.Name)
	_, err := server.platform.Droplets.Delete(server.ctx, server.droplet.ID)
	if err != nil {
		return errors.Wrap(err, "failed to delete droplet")
	}

	server.droplet = nil
	return nil
}
