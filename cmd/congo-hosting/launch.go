package main

import (
	"flag"
	"os"

	"github.com/ccutch/congo/pkg/congo_host"
)

func launch(args ...string) (server *congo_host.Server, err error) {
	var (
		cmd     = flag.NewFlagSet("launch", flag.ExitOnError)
		apiKey  = cmd.String("api-key", "$DIGITAL_OCEAN_API_KEY", "Digital Ocean API Key")
		path    = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		name    = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		region  = cmd.String("region", "sfo2", "Region of Digital Ocean droplet")
		size    = cmd.String("size", "s-1vcpu-2gb", "Size of Digital Ocean droplet")
		storage = cmd.Int64("storage", 5, "Volume size of Digital Ocean droplet")
	)

	if err := cmd.Parse(args[1:]); err != nil {
		return nil, err
	}

	if *apiKey == "$DIGITAL_OCEAN_API_KEY" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	host := congo_host.InitCongoHost(*path, congo_host.WithApiToken(*apiKey))
	return host.NewServer(*name, *region, *size, *storage)
}
