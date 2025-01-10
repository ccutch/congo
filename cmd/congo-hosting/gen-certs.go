package main

import (
	"flag"
	"log"
	"os"

	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_host/platforms/digitalocean"
	"github.com/pkg/errors"
)

func genCerts(args ...string) error {
	var (
		cmd        = flag.NewFlagSet("gen-certs", flag.ExitOnError)
		apiKey     = cmd.String("api-key", "$DIGITAL_OCEAN_API_KEY", "Digital Ocean API Key")
		path       = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		name       = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		domainName = cmd.String("domain", "", "Domain name to generate cert for")
	)

	if err := cmd.Parse(args[1:]); err != nil {
		return err
	}

	if *apiKey == "$DIGITAL_OCEAN_API_KEY" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	if *domainName == "" {
		log.Fatal("Missing domain name")
	}

	host := congo_host.InitCongoHost(*path, digitalocean.NewClient(*apiKey))
	server, err := host.GetServer(*name)
	if err != nil {
		return errors.Wrap(err, "failed to get server")
	}

	if err := server.Reload(); err != nil {
		return errors.Wrap(err, "failed to reload server")
	} else if domain, err := server.NewDomain(*domainName); err != nil {
		return errors.Wrap(err, "failed to create domain")
	} else if err := domain.Verify(); err != nil {
		return errors.Wrap(err, "failed to verify domain")
	} else {
		return server.Restart()
	}
}
