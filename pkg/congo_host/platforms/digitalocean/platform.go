package digitalocean

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Client struct {
	*godo.Client
	host *congo_host.CongoHost
}

func NewClient(token string) congo_host.Platform {
	if token == "" {
		return nil
	}
	return &Client{godo.NewClient(oauth2.NewClient(
		context.Background(),
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
	)), nil}
}

func (client *Client) Init(host *congo_host.CongoHost) {
	client.host = host
}

func (client *Client) Server(name string) congo_host.Server {
	return &Server{client: client, Name: name}
}

type Server struct {
	client *Client
	Name   string
	IP     string
	sshKey *godo.Key
	volume *godo.Volume
}

func (d *Server) Launch(region string, size string, storage int64) error {
	if err := d.setupAccess(); err != nil {
		return errors.Wrap(err, "failed to create droplet")
	}
	if err := d.setupVolume(region, storage); err != nil {
		return errors.Wrap(err, "failed to create droplet")
	}
	if err := d.createDroplet(region, size); err != nil {
		return errors.Wrap(err, "failed to create droplet")
	}
	return nil
}

func (d *Server) Delete(purge bool, force bool) error {
	if err := d.deleteDroplet(); !force && err != nil {
		return errors.Wrap(err, "failed to delete droplet")
	}
	if err := d.deleteRemoteKeys(); !force && err != nil {
		return errors.Wrap(err, "failed to delete remote keys")
	}
	if err := d.deleteLocalKeys(); !force && err != nil {
		return errors.Wrap(err, "failed to delete local keys")
	}
	if purge {
		time.Sleep(15 * time.Second)
		if err := d.deleteVolume(); !force && err != nil {
			return errors.Wrap(err, "failed to delete volume")
		}
	}
	return nil
}

func (d *Server) Reload() (err error) {
	var (
		droplet  *godo.Droplet
		droplets []godo.Droplet
		volumes  []godo.Volume
		keys     []godo.Key
	)

	if droplets, _, err = d.client.Droplets.ListByName(context.TODO(), d.Name, nil); err != nil {
		return errors.Wrap(err, "failed to list droplets")
	}

	if len(droplets) == 1 {
		droplet = &droplets[0]
	} else {
		return errors.New("no droplet found")
	}

	d.Name = droplet.Name
	if d.IP, err = droplet.PublicIPv4(); err != nil {
		return errors.Wrap(err, "failed to get droplet IP")
	}

	opt := &godo.ListVolumeParams{
		Name:   d.Name + "-data",
		Region: droplet.Region.Slug,
	}

	if volumes, _, err = d.client.Storage.ListVolumes(context.TODO(), opt); err != nil {
		return errors.Wrap(err, "failed to list volumes")
	}

	if len(volumes) == 1 {
		d.volume = &volumes[0]
	}

	if keys, _, err = d.client.Keys.List(context.TODO(), nil); err != nil {
		return errors.Wrap(err, "failed to list keys")
	}

	for _, key := range keys {
		if key.Name == d.Name+"-admin-key" {
			d.sshKey = &key
			break
		}
	}

	return nil
}

func (s *Server) Addr() string {
	return s.IP
}

func (s *Server) Run(stdin io.Reader, stdout io.Writer, args ...string) error {
	if err := s.Reload(); err != nil {
		return errors.Wrap(err, "failed to reload server")
	}
	var stderr bytes.Buffer
	_, priKey := s.keys()
	cmd := exec.Command(
		"ssh", "-i", priKey,
		"-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", s.IP),
		strings.Join(args, " "))
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = &stderr
	return errors.Wrap(cmd.Run(), strings.Join(args, " ")+"\n"+stderr.String())
}

func (s *Server) Copy(source, dest string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	_, priKey := s.keys()
	cmd := exec.Command(
		"scp",
		"-o", "StrictHostKeyChecking=no",
		"-i", priKey,
		source,
		fmt.Sprintf("root@%s:%s", s.IP, dest))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return stdout, stderr, cmd.Run()
}

func (s *Server) Assign(domain *congo_host.Domain) error {
	ctx, parts := context.Background(), strings.SplitN(domain.ID, ".", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid domain, expecting subdomain.topleveldomain.extension")
	}

	subdomain, rootDomain := parts[0], parts[1]
	records, _, err := s.client.Domains.Records(ctx, rootDomain, &godo.ListOptions{})
	if err != nil {
		return err
	}

	dnsRecordExists := false
	for _, record := range records {
		if record.Type == "A" && record.Name == subdomain && record.Data == s.IP {
			log.Println("DNS record already exists for domain:", domain.ID)
			dnsRecordExists = true
			break
		}
	}

	if !dnsRecordExists {
		_, _, err = s.client.Domains.CreateRecord(ctx, rootDomain, &godo.DomainRecordEditRequest{
			Type: "A",
			Name: subdomain,
			Data: s.IP,
			TTL:  3600,
		})

		if err != nil {
			return err
		}

		_, _, err = s.client.Domains.CreateRecord(ctx, rootDomain, &godo.DomainRecordEditRequest{
			Type: "A",
			Name: "*." + subdomain,
			Data: s.IP,
			TTL:  3600,
		})

		if err != nil {
			return err
		}

		log.Println("Created A record for domain:", domain.ID, "with IP:", s.IP)
		log.Println("Created A record for domain:", "*."+domain.ID, "with IP:", s.IP)
	}

	return nil
}

func (s *Server) Remove(domain *congo_host.Domain) error {
	ctx, parts := context.Background(), strings.SplitN(domain.ID, ".", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid domain, expecting subdomain.topleveldomain.extension")
	}

	subdomain, rootDomain := parts[0], parts[1]
	records, _, err := s.client.Domains.Records(ctx, rootDomain, &godo.ListOptions{})
	if err != nil {
		return err
	}

	dnsRecordExists := false
	for _, record := range records {
		if record.Type == "A" && record.Name == subdomain && record.Data == s.IP {
			dnsRecordExists = true
			break
		}
	}

	if !dnsRecordExists {
		return nil
	}

	if _, err = s.client.Domains.Delete(ctx, domain.ID); err != nil {
		return err
	}

	if _, err = s.client.Domains.Delete(ctx, "*."+domain.ID); err != nil {
		return err
	}

	log.Println("Deleted A record for domain:", domain.ID, "with IP:", s.IP)
	log.Println("Deleted A record for domain:", "*."+domain.ID, "with IP:", s.IP)

	return nil
}
