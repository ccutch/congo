package hosting

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/digitalocean/godo"
)

type Server struct {
	*Client
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
}

func (client *Client) NewServer(name, region, size string, storage int64) (*Server, error) {
	server := Server{Client: client, Name: name, Region: region, ctx: context.Background()}
	server.setupAccessKey()
	server.setupVolumne(storage)
	server.startDroplet(size)
	server.Start()
	return &server, server.Err
}

func (client *Client) LoadServer(name, region string) (*Server, error) {
	server := Server{Client: client, Name: name, Region: region, ctx: context.Background()}
	server.pubKey = fmt.Sprintf("%s/id_rsa.pub", filepath.Join(client.dataPath, name))
	server.priKey = fmt.Sprintf("%s/id_rsa", filepath.Join(client.dataPath, name))
	server.Refresh()
	return &server, server.Err
}

func (server *Server) Refresh() {
	var droplets []godo.Droplet
	droplets, _, server.Err = server.platform.Droplets.ListByName(server.ctx, server.Name, nil)
	if len(droplets) == 1 {
		server.droplet = &droplets[0]
		server.IP, server.Err = server.droplet.PublicIPv4()
	}
}

func (server *Server) Setup() {
	server.Err = server.Run(fmt.Sprintf(mountVolume, server.Name+"-data"))
	server.Err = server.Run(installGolang)
	server.Err = server.Run(setupFirewall)
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
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (server *Server) Start() {
	server.Err = exec.Command("go", "build", "-o", "congo", ".").Run()
	server.Err = server.Copy("./congo", "/root/congo")
	time.Sleep(2 * time.Second)
	server.Err = server.Run(fmt.Sprintf(startServer, 80))
}

func (server *Server) GenerateCerts(domain string) {
	server.Err = server.Run(fmt.Sprintf(generateCerts, domain))
}

func (server *Server) setupAccessKey() {
	if server.Err != nil {
		return
	}
	server.pubKey, server.priKey, server.Err = server.GenerateSSHKey(server.Name)
	if server.Err != nil {
		return
	}
	var data []byte
	data, server.Err = os.ReadFile(server.pubKey)
	if server.Err != nil {
		return
	}
	server.sshKey, _, server.Err = server.Client.platform.Keys.Create(server.ctx, &godo.KeyCreateRequest{
		Name:      "congo-admin-key",
		PublicKey: string(data),
	})
}

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
