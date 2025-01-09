package digitalocean

import (
	"context"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

func (d *Server) setupVolume(region string, size int64) (err error) {
	d.volume, _, err = d.client.Storage.CreateVolume(context.TODO(), &godo.VolumeCreateRequest{
		Name:          d.Name + "-data",
		Region:        region,
		SizeGigaBytes: size,
		Description:   "volume for congo server",
	})
	return err
}

func (server *Server) deleteVolume() error {
	_, err := server.client.Storage.DeleteVolume(context.TODO(), server.volume.ID)
	server.volume = nil
	return errors.Wrap(err, "failed to delete volume")
}
