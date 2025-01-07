package congo_host

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

func (server *Server) setupVolumne(size int64) {
	if server.Error != nil {
		return
	}
	ctx := context.Background()
	server.volume, _, server.Error = server.host.platform.Storage.CreateVolume(ctx, &godo.VolumeCreateRequest{
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

	ctx := context.Background()
	fmt.Printf("Deleting volume %s...\n", server.volume.Name)
	_, err := server.host.platform.Storage.DeleteVolume(ctx, server.volume.ID)
	if err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	server.volume = nil
	return nil
}
