package main

import (
	"flag"
	"os"

	"github.com/ccutch/congo/internal/hosting"
)

func launch(args ...string) (server *hosting.Server, err error) {
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

	client := hosting.NewClient(*path, *apiKey)
	return client.NewServer(*name, *region, *size, *storage)
}
