package congo_host

import (
	"fmt"

	"github.com/digitalocean/godo"
)

func (server *Server) setupVolumne(size int64) {
	if server.Err != nil {
		return
	}
	server.volume, _, server.Err = server.platform.Storage.CreateVolume(server.ctx, &godo.VolumeCreateRequest{
		Name:          server.Name + "-data",
		Region:        server.Region,
		SizeGigaBytes: size,
		Description:   "volume for congo server",
	})
}

func (server *Server) deleteVolume() error {
	if server.volume == nil {
		return nil
	}

	fmt.Printf("Deleting volume %s...\n", server.volume.Name)
	_, err := server.platform.Storage.DeleteVolume(server.ctx, server.volume.ID)
	if err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	server.volume = nil
	return nil
}
