package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ccutch/congo/pkg/hosting"
)

func destroy(args ...string) error {
	var (
		cmd    = flag.NewFlagSet("destroy", flag.ExitOnError)
		apiKey = cmd.String("api-key", "$DIGITAL_OCEAN_API_KEY", "Digital Ocean API Key")
		path   = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		name   = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		region = cmd.String("region", "sfo2", "Region of Digital Ocean droplet")
	)

	if err := cmd.Parse(args[1:]); err != nil {
		return err
	}

	if *apiKey == "$DIGITAL_OCEAN_API_KEY" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	client := hosting.NewClient(*path, *apiKey)

	server, err := client.LoadServer(*name, *region)
	if err != nil {
		return fmt.Errorf("failed to load server: %w", err)
	}

	if err := server.Destroy(); err != nil {
		return fmt.Errorf("failed to destroy server: %w", err)
	}

	fmt.Printf("Server %s destroyed successfully.\n", *name)
	return nil
}
