package main

import (
	"flag"
	"log"
	"os"
	"os/exec"

	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_host/platforms/digitalocean"
	"github.com/pkg/errors"
)

func launch(args ...string) (*congo_host.RemoteHost, error) {
	var (
		cmd     = flag.NewFlagSet("launch", flag.ExitOnError)
		apiKey  = cmd.String("api-key", "$DIGITAL_OCEAN_API_KEY", "Digital Ocean API Key")
		path    = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		app     = cmd.String("app", "", "Applicationto use for the server")
		web     = cmd.String("web", "", "Website we are deploying")
		name    = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		region  = cmd.String("region", "sfo2", "Region of Digital Ocean droplet")
		size    = cmd.String("size", "s-1vcpu-2gb", "Size of Digital Ocean droplet")
		storage = cmd.Int64("storage", 5, "Volume size of Digital Ocean droplet")
		domain  = cmd.String("domain", "", "Domain name to generate cert for")
		enVars  stringArray
	)

	cmd.Var(&enVars, "env", "Environment variables to include in environment")
	if err := cmd.Parse(args[1:]); err != nil {
		return nil, err
	}

	if *apiKey == "$DIGITAL_OCEAN_API_KEY" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	log.Println("Recording record...")
	host := congo_host.InitCongoHost(*path,
		congo_host.WithPlatform(digitalocean.NewClient(*apiKey)))
	server, err := host.NewServer(*name, *size, *region)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create server")
	}

	log.Println("Launching server...")
	if err = server.Launch(*region, *size, *storage); err != nil {
		return nil, errors.Wrap(err, "failed to create server")
	}

	if *domain != "" {
		domain := server.Domain(*domain)
		if err = server.Assign(domain); err != nil {
			return nil, errors.Wrap(err, "failed to assign domain")
		}

		if err = domain.Save(); err != nil {
			return nil, errors.Wrap(err, "failed to save domain")
		}

		if err = domain.Verify("admin@" + domain.DomainName); err != nil {
			return nil, errors.Wrap(err, "failed to verify domain")
		}
	}

	if *app != "" {
		log.Println("Building binary...")
		if err = exec.Command("go", "build", "-o", "congo", "./apps/"+*app).Run(); err != nil {
			log.Println("Failed to build binary: ", err)
			return nil, errors.Wrap(err, "failed to build binary")
		}
	} else if *web != "" {
		log.Println("Building website...")
		if err = exec.Command("go", "build", "-o", "congo", "./web/"+*web).Run(); err != nil {
			log.Println("Failed to build website: ", err)
			return nil, errors.Wrap(err, "failed to build website")
		}
	}

	log.Println("Deploying binary...")
	if err = server.Deploy("congo"); err != nil {
		return nil, errors.Wrap(err, "failed to deploy binary")
	}

	return server, err
}
