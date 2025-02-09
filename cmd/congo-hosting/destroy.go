package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_host/platforms/digitalocean"
	"github.com/pkg/errors"
)

func destroy(args ...string) error {
	var (
		cmd    = flag.NewFlagSet("destroy", flag.ExitOnError)
		apiKey = cmd.String("api-key", "", "Digital Ocean API Key default to environ")
		path   = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		name   = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		purge  = cmd.Bool("purge", false, "Destroy droplet and purge data volumes")
		force  = cmd.Bool("force", false, "Force destroy even if there are errors")
	)

	if err := cmd.Parse(args[1:]); err != nil {
		return err
	}

	if *apiKey == "" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	host := congo_host.InitCongoHost(*path,
		congo_host.WithPlatform(digitalocean.NewClient(*apiKey)))
	server, err := host.GetServer(*name)
	if err != nil {
		return errors.Wrap(err, "failed to get server")
	}

	if err = server.Reload(); !*force && err != nil {
		return errors.Wrap(err, "failed to load server")
	}

	domains, err := server.Domains()
	log.Println("deleting domains", domains)
	if err != nil {
		return errors.Wrap(err, "failed to get domains")
	}

	for _, domain := range domains {
		log.Println("deleting domain", domain.ID)
		if err = server.Remove(domain); err != nil {
			return errors.Wrap(err, "failed to remove domain")
		}
		if err = domain.Delete(); err != nil {
			return errors.Wrap(err, "failed to delete domain")
		}
	}

	if err = server.Delete(*purge, *force); err != nil {
		return errors.Wrap(err, "failed to destroy server")
	}

	fmt.Printf("Server %s destroyed successfully.\n", *name)
	return nil
}
