package digitalocean

import (
	"context"
	_ "embed"
	"time"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

func (s *Server) createDroplet(region, size string) error {
	_, _, err := s.client.Droplets.Create(context.TODO(), &godo.DropletCreateRequest{
		Name:   s.Name,
		Region: region,
		Size:   size,
		Image: godo.DropletCreateImage{
			Slug: "docker-20-04",
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{Fingerprint: s.sshKey.Fingerprint},
		},
		Volumes: []godo.DropletCreateVolume{
			{ID: s.volume.ID},
		},
		Backups: true,
	})

	if err != nil {
		return err
	}

	for s.IP == "" {
		time.Sleep(10 * time.Second)
		s.Reload()
	}

	time.Sleep(30 * time.Second)
	return err
}

func (s *Server) deleteDroplet() error {
	droplets, _, err := s.client.Droplets.ListByName(context.TODO(), s.Name, nil)
	if err != nil {
		return errors.Wrap(err, "failed to list droplets")
	}

	if len(droplets) == 0 {
		return errors.New("no droplet found")
	} else if len(droplets) > 1 {
		return errors.New("multiple droplets found")
	}

	_, err = s.client.Droplets.Delete(context.TODO(), droplets[0].ID)
	return errors.Wrap(err, "failed to delete droplet")
}
