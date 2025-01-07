package congo_host

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

type Server struct {
	*CongoHost
	Name   string
	Region string
	IP     string
	Error  error

	sshKey  *godo.Key
	volume  *godo.Volume
	droplet *godo.Droplet

	ctx    context.Context
	pubKey string
	priKey string

	Stdin  io.Reader
	Stdout io.Writer
}

func (host *CongoHost) NewServer(name, region, size string, storage int64) (*Server, error) {
	server := Server{CongoHost: host, Name: name, Region: region, ctx: context.Background(), Stdin: os.Stdin, Stdout: os.Stdout}
	server.setupAccessKey()
	server.setupVolumne(storage)
	server.startDroplet(size)
	server.prepareServer()
	return &server, server.Error
}

func (host *CongoHost) LoadServer(name, region string) (*Server, error) {
	server := Server{CongoHost: host, Name: name, Region: region, ctx: context.Background(), Stdin: os.Stdin, Stdout: os.Stdout}
	server.pubKey = fmt.Sprintf("%s/id_rsa.pub", filepath.Join(host.root, "hosts", name))
	server.priKey = fmt.Sprintf("%s/id_rsa", filepath.Join(host.root, "hosts", name))
	server.checkAccessKeys()
	server.Refresh()
	return &server, server.Error
}

func (host *CongoHost) ListServers() ([]*Server, error) {
	var servers []*Server
	if _, err := os.Stat(filepath.Join(host.root, "hosts")); os.IsNotExist(err) {
		return []*Server{}, nil
	}

	entries, err := os.ReadDir(filepath.Join(host.root, "hosts"))
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			server := Server{
				CongoHost: host,
				Name:      entry.Name(),
				ctx:       context.Background()}
			servers = append(servers, &server)
		}
	}

	return servers, nil
}

//go:embed resources/start-server.sh
var startServer string

func (server *Server) Start() {
	if server.Error != nil {
		return
	}

	server.Error = server.Run(fmt.Sprintf(startServer, 8080))
}

func (server *Server) Refresh() {
	var (
		droplets []godo.Droplet
		volumes  []godo.Volume
		keys     []godo.Key
	)

	if droplets, _, server.Error = server.platform.Droplets.ListByName(server.ctx, server.Name, nil); server.Error != nil {
		server.Error = errors.Wrap(server.Error, "failed to list droplets")
		return
	}

	if len(droplets) == 1 {
		server.droplet = &droplets[0]
		server.IP, server.Error = server.droplet.PublicIPv4()
	}

	opt := &godo.ListVolumeParams{
		Name:   server.Name + "-data",
		Region: server.Region,
	}

	if volumes, _, server.Error = server.platform.Storage.ListVolumes(server.ctx, opt); server.Error != nil {
		server.Error = errors.Wrap(server.Error, "failed to list volumes")
		return
	}

	if len(volumes) == 1 {
		server.volume = &volumes[0]
	}

	if keys, _, server.Error = server.platform.Keys.List(server.ctx, nil); server.Error != nil {
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

	return nil
}

func (server *Server) Purge(force bool) error {
	if err := server.deleteVolume(); !force && err != nil {
		return errors.Wrap(err, "failed to delete volume")
	}

	return nil
}
