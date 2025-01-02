package congo_host

import (
	"bytes"
	"cmp"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

type Server struct {
	*CongoHost
	Name   string
	Region string
	IP     string
	Err    error

	sshKey  *godo.Key
	volume  *godo.Volume
	droplet *godo.Droplet

	ctx    context.Context
	pubKey string
	priKey string

	Stdin  io.Reader
	Stdout io.Writer
}

func (client *CongoHost) NewServer(name, region, size string, storage int64) (*Server, error) {
	server := Server{CongoHost: client, Name: name, Region: region, ctx: context.Background(), Stdin: os.Stdin, Stdout: os.Stdout}
	server.setupAccessKey()
	server.setupVolumne(storage)
	server.startDroplet(size)
	server.setupService()
	return &server, server.Err
}

func (client *CongoHost) LoadServer(name, region string) (*Server, error) {
	server := Server{CongoHost: client, Name: name, Region: region, ctx: context.Background(), Stdin: os.Stdin, Stdout: os.Stdout}
	server.pubKey = fmt.Sprintf("%s/id_rsa.pub", filepath.Join(client.root, name))
	server.priKey = fmt.Sprintf("%s/id_rsa", filepath.Join(client.root, name))
	server.checkAccessKeys()
	server.Refresh()
	return &server, server.Err
}

func (host *CongoHost) ListServers() ([]*Server, error) {
	var servers []*Server
	if _, err := os.Stat(host.root); os.IsNotExist(err) {
		return []*Server{}, nil
	}
	entries, err := os.ReadDir(host.root)
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

func (server *Server) Refresh() {
	var (
		droplets []godo.Droplet
		volumes  []godo.Volume
		keys     []godo.Key
	)
	if droplets, _, server.Err = server.platform.Droplets.ListByName(server.ctx, server.Name, nil); server.Err != nil {
		return
	}
	if len(droplets) == 1 {
		server.droplet = &droplets[0]
		server.IP, server.Err = server.droplet.PublicIPv4()
	}
	opt := &godo.ListVolumeParams{
		Name:   server.Name + "-data",
		Region: server.Region,
	}
	if volumes, _, server.Err = server.platform.Storage.ListVolumes(server.ctx, opt); server.Err != nil {
		return
	}
	if len(volumes) == 1 {
		server.volume = &volumes[0]
	}
	if keys, _, server.Err = server.platform.Keys.List(server.ctx, nil); server.Err != nil {
		return
	}
	for _, key := range keys {
		if key.Name == server.Name+"-admin-key" {
			server.sshKey = &key
			break
		}
	}
}

func (server *Server) Run(args ...string) error {
	if server.Err != nil {
		return server.Err
	}
	cmd := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-i", server.priKey,
		fmt.Sprintf("root@%s", server.IP),
		strings.Join(args, " "),
	)
	var buf bytes.Buffer
	cmd.Stdin = server.Stdin
	cmd.Stdout = server.Stdout
	cmd.Stderr = &buf
	return errors.Wrap(cmd.Run(), buf.String())
}

func (server *Server) Copy(source, dest string) error {
	if server.Err != nil {
		return server.Err
	}
	cmd := exec.Command(
		"scp",
		"-o", "StrictHostKeyChecking=no",
		"-i", server.priKey,
		source,
		fmt.Sprintf("root@%s:%s", server.IP, dest),
	)
	var buf bytes.Buffer
	cmd.Stdin = cmp.Or[io.Reader](server.Stdin, os.Stdin)
	cmd.Stdout = cmp.Or[io.Writer](server.Stdout, os.Stdout)
	cmd.Stderr = &buf
	return errors.Wrap(cmd.Run(), buf.String())
}

func (server *Server) Deploy(source string) error {
	if server.Err == nil {
		server.Err = server.Copy(source, "/root/congo")
	}

	if server.Err == nil {
		server.Start()
	}

	return server.Err
}

//go:embed resources/start-server.sh
var startServer string

func (server *Server) Start() {
	if server.Err != nil {
		return
	}
	server.Err = server.Run(fmt.Sprintf(startServer, 8080))
}

func (server *Server) startDroplet(size string) {
	if server.Err != nil {
		return
	}
	server.droplet, _, server.Err = server.platform.Droplets.Create(server.ctx, &godo.DropletCreateRequest{
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
	if server.Err != nil {
		return
	}
	for server.IP == "" {
		fmt.Printf("Server Info: %v\n", server)
		time.Sleep(10 * time.Second)
		if server.Refresh(); server.Err != nil {
			log.Fatalf("Error checking status: %s", server.Err)
		}
	}
	time.Sleep(30 * time.Second)
}

//go:embed resources/setup-server.sh
var setupServer string

func (server *Server) setupService() {
	if server.Err != nil {
		return
	}
	server.Err = server.Run(fmt.Sprintf(setupServer, server.Name+"-data"))
}

func (server *Server) Destroy(force bool) error {
	if server.Err != nil {
		return server.Err
	}

	if err := server.deleteDroplet(); !force && err != nil {
		return err
	} else {
		time.Sleep(15 * time.Second)
	}

	if err := server.deleteVolume(); !force && err != nil {
		return err
	}

	if err := server.deleteSSHKey(); !force && err != nil {
		return err
	}

	if err := server.deleteLocalKeys(); !force && err != nil {
		return err
	}

	fmt.Println("Server resources destroyed successfully.")
	return nil

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
