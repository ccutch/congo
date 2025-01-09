package congo_host

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

type Server struct {
	congo.Model
	local      bool
	Name       string
	Region     string
	Size       string
	VolumeSize int64
	IP         string
	Error      error

	host    *CongoHost
	sshKey  *godo.Key
	volume  *godo.Volume
	droplet *godo.Droplet

	Stdin  io.Reader
	Stdout io.Writer
}

func (host *CongoHost) LocalHost() *Server {
	return &Server{host: host, local: true, Model: host.db.NewModel("local"), Stdin: os.Stdin, Stdout: os.Stdout}
}

func (host *CongoHost) Server(name string) *Server {
	return &Server{host: host, Model: host.db.NewModel(name), Name: name, Stdin: os.Stdin, Stdout: os.Stdout}
}

func (s *Server) Create(region, size string, storage int64) error {
	s.Region, s.Size, s.VolumeSize = region, size, storage
	return s.host.db.Query(`

	INSERT INTO servers (id, name, region, size, volume_size)
	VALUES (?, ?, ?, ?, ?)
	RETURNING created_at, updated_at

	`, s.ID, s.Name, s.Region, s.Size, s.VolumeSize).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (server *Server) Load() error {
	err := ""
	server.Error = server.host.db.Query(`

		SELECT id, name, region, size, volume_size, ip_address, error, created_at, updated_at
		FROM servers
		WHERE name = ?

	`, server.Name).Scan(&server.ID, &server.Name, &server.Region, &server.Size, &server.VolumeSize, &server.IP, &err, &server.CreatedAt, &server.UpdatedAt)
	if server.Error != nil {
		return server.Error
	}
	if err != "" {
		server.Error = errors.New(err)
	}
	server.checkAccessKeys()
	server.Refresh()
	return server.Error
}

func (host *CongoHost) ListServers() (servers []*Server, err error) {
	return servers, host.db.Query(`

		SELECT id, name, region, size, volume_size, ip_address, error, created_at, updated_at
		FROM servers
		ORDER BY created_at DESC

	`).All(func(scan congo.Scanner) error {
		s, sErr := Server{Model: host.db.Model()}, ""
		if err = scan(&s.ID, &s.Name, &s.Region, &s.Size, &s.VolumeSize, &s.IP, &sErr, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return err
		}

		if sErr != "" {
			s.Error = errors.New(sErr)
		}

		go func() {
			s.checkAccessKeys()
			s.Refresh()
		}()

		servers = append(servers, &s)
		return nil
	})
}

func (s *Server) Keys() (string, string) {
	pubKey := fmt.Sprintf("%s/id_rsa.pub", filepath.Join(s.host.db.Root, "hosts", s.Name))
	priKey := fmt.Sprintf("%s/id_rsa", filepath.Join(s.host.db.Root, "hosts", s.Name))
	return pubKey, priKey
}

//go:embed resources/server/start-server.sh
var startServer string

func (s *Server) Setup() {
	if s.Error != nil {
		return
	}
	s.setupAccessKey()
	s.setupVolumne()
	s.startDroplet()
	s.prepareServer()
}

func (s *Server) Start() {
	if s.Error != nil {
		return
	}
	s.Error = s.Run(fmt.Sprintf(startServer, 8080))
}

func (server *Server) Refresh() {
	var (
		droplets []godo.Droplet
		volumes  []godo.Volume
		keys     []godo.Key
	)

	ctx := context.Background()
	if droplets, _, server.Error = server.host.platform.Droplets.ListByName(ctx, server.Name, nil); server.Error != nil {
		server.Error = errors.Wrap(server.Error, "failed to list droplets")
		return
	}

	if len(droplets) == 1 {
		server.droplet = &droplets[0]
		server.IP, server.Error = server.droplet.PublicIPv4()
		if server.Error = server.Save(); server.Error != nil {
			return
		}
	}

	opt := &godo.ListVolumeParams{
		Name:   server.Name + "-data",
		Region: server.Region,
	}

	if volumes, _, server.Error = server.host.platform.Storage.ListVolumes(ctx, opt); server.Error != nil {
		server.Error = errors.Wrap(server.Error, "failed to list volumes")
		return
	}

	if len(volumes) == 1 {
		server.volume = &volumes[0]
	}

	if keys, _, server.Error = server.host.platform.Keys.List(ctx, nil); server.Error != nil {
		server.Error = errors.Wrap(server.Error, "failed to list keys")
		return
	}

	for _, key := range keys {
		if key.Name == server.Name+"-admin-key" {
			server.sshKey = &key
			break
		}
	}
}

func (server *Server) Destroy(force bool) error {
	if server.Error != nil {
		return server.Error
	}

	if err := server.deleteDroplet(); !force && err != nil {
		return errors.Wrap(err, "failed to delete droplet")
	}

	if err := server.deleteRemoteKeys(); !force && err != nil {
		return errors.Wrap(err, "failed to delete remote keys")
	}

	if err := server.deleteLocalKeys(); !force && err != nil {
		return errors.Wrap(err, "failed to delete local keys")
	}

	return errors.Wrap(server.Delete(), "failed to delete server")
}

func (server *Server) Purge(force bool) error {
	if err := server.deleteVolume(); !force && err != nil {
		return errors.Wrap(err, "failed to delete volume")
	}

	return nil
}

func (server *Server) Save() error {
	var err string
	if server.Error != nil {
		err = server.Error.Error()
	}
	return server.host.db.Query(`
	
		UPDATE servers
		SET name = ?,
				region = ?,
				size = ?,
				volume_size = ?,
				ip_address = ?,
				error = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, server.Name, server.Region, server.Size, server.VolumeSize, server.IP, err, server.ID).Scan(&server.UpdatedAt)
}

func (server *Server) Delete() error {
	return server.host.db.Query(`
	
		DELETE FROM servers
		WHERE id = ?

	`, server.ID).Exec()
}
