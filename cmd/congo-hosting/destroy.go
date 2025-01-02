package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ccutch/congo/pkg/congo_host"
)

func destroy(args ...string) error {
	var (
		cmd    = flag.NewFlagSet("destroy", flag.ExitOnError)
		apiKey = cmd.String("api-key", "$DIGITAL_OCEAN_API_KEY", "Digital Ocean API Key")
		path   = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		name   = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		region = cmd.String("region", "sfo2", "Region of Digital Ocean droplet")
		force  = cmd.Bool("force", false, "Force destroy even if there are errors")
	)

	if err := cmd.Parse(args[1:]); err != nil {
		return err
	}

	if *apiKey == "$DIGITAL_OCEAN_API_KEY" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	host := congo_host.InitCongoHost(*path, congo_host.WithApiToken(*apiKey))
	server, err := host.LoadServer(*name, *region)
	if err != nil {
		return fmt.Errorf("failed to load server: %w", err)
	}

	if err := server.Destroy(*force); err != nil {
		return fmt.Errorf("failed to destroy server: %w", err)
	}

	fmt.Printf("Server %s destroyed successfully.\n", *name)
	return nil
}
