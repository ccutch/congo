package congo_host

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
	server.setupService()
	return &server, server.Err
}

func (client *Client) LoadServer(name, region string) (*Server, error) {
	server := Server{Client: client, Name: name, Region: region, ctx: context.Background()}
	server.pubKey = fmt.Sprintf("%s/id_rsa.pub", filepath.Join(client.dataPath, name))
	server.priKey = fmt.Sprintf("%s/id_rsa", filepath.Join(client.dataPath, name))
	server.checkAccessKeys()
	server.Refresh()
	return &server, server.Err
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
	server.Err = server.Run(fmt.Sprintf(startServer, 8080))
}

func (server *Server) GenerateCerts(domain string) {
	server.Err = server.Run(fmt.Sprintf(generateCerts, domain))
}

func (server *Server) checkAccessKeys() {
	if server.Err != nil {
		return
	}
	if _, server.Err = os.Stat(server.pubKey); server.Err != nil {
		return
	}
	_, server.Err = os.Stat(server.priKey)
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
		Name:      server.Name + "-admin-key",
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

func (server *Server) setupService() {
	if server.Err != nil {
		return
	}
	server.Err = server.Run(fmt.Sprintf(setupServer, server.Name+"-data"))
	exec.Command("go", "build", "-o", "congo", ".").Run()
	if server.Err = server.Copy("./congo", "/root/congo"); server.Err != nil {
		log.Fatal("Failed to load congo binary", server.Err)
	}
	server.Start()
}
func (server *Server) Destroy() error {
	if server.Err != nil {
		return server.Err
	}

	if err := server.deleteDroplet(); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	if err := server.deleteVolume(); err != nil {
		return err
	}

	if err := server.deleteSSHKey(); err != nil {
		return err
	}

	if err := server.deleteLocalKeys(); err != nil {
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

func (server *Server) deleteSSHKey() error {
	if server.sshKey == nil {
		return nil
	}

	fmt.Printf("Deleting SSH key %s...\n", server.sshKey.Name)
	_, err := server.platform.Keys.DeleteByID(server.ctx, server.sshKey.ID)
	if err != nil {
		return fmt.Errorf("failed to delete SSH key: %w", err)
	}

	server.sshKey = nil
	return nil
}

func (server *Server) deleteLocalKeys() error {

	if server.priKey != "" {
		if err := os.Remove(server.priKey); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete private key file: %w", err)
		}
	}

	if server.pubKey != "" {
		if err := os.Remove(server.pubKey); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete public key file: %w", err)
		}
	}

	return nil
}
