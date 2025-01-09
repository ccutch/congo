package main

import (
	"flag"
	"log"
	"os"
	"os/exec"

	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/pkg/errors"
)

func launch(args ...string) (*congo_host.Server, error) {
	var (
		cmd     = flag.NewFlagSet("launch", flag.ExitOnError)
		apiKey  = cmd.String("api-key", "$DIGITAL_OCEAN_API_KEY", "Digital Ocean API Key")
		path    = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		app     = cmd.String("app", "", "Prototype to use for the server")
		name    = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		region  = cmd.String("region", "sfo2", "Region of Digital Ocean droplet")
		size    = cmd.String("size", "s-1vcpu-2gb", "Size of Digital Ocean droplet")
		storage = cmd.Int64("storage", 5, "Volume size of Digital Ocean droplet")
		domain  = cmd.String("domain", "", "Domain name to generate cert for")
	)

	if err := cmd.Parse(args[1:]); err != nil {
		return nil, err
	}

	if *app == "" {
		return nil, errors.New("Choose app: blogfront, launchpad, workbench")
	}

	if *apiKey == "$DIGITAL_OCEAN_API_KEY" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	host := congo_host.InitCongoHost(*path, congo_host.WithApiToken(*apiKey))
	server := host.Server(*name)
	err := server.Create(*region, *size, *storage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create server")
	}

	server.Setup()
	if err = exec.Command("go", "build", "-o", "congo", "./apps/"+*app).Run(); err != nil {
		log.Println("Failed to build binary: ", err)
		return nil, errors.Wrap(err, "failed to build binary")
	}

	if err = server.Deploy("congo"); err != nil {
		return nil, errors.Wrap(err, "failed to deploy binary")
	}

	if *domain != "" {
		log.Printf("Add A record for %s to %s", *domain, server.IP)
		if d, err := server.NewDomain(*domain); err != nil {
			return nil, errors.Wrap(err, "failed to create domain")
		} else if err = d.Verify(); err != nil {
			return nil, errors.Wrap(err, "failed to verify domain")
		}
	}

	return server, err
}
