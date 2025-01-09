package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ccutch/congo/pkg/congo_host"
)

func destroy(args ...string) error {
	var (
		cmd    = flag.NewFlagSet("destroy", flag.ExitOnError)
		apiKey = cmd.String("api-key", "", "Digital Ocean API Key default to environ")
		path   = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		name   = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		force  = cmd.Bool("force", false, "Force destroy even if there are errors")
		purge  = cmd.Bool("purge", false, "Destroy droplet and purge data volumes")
	)

	if err := cmd.Parse(args[1:]); err != nil {
		return err
	}

	if *apiKey == "" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	host := congo_host.InitCongoHost(*path, congo_host.WithApiToken(*apiKey))
	server := host.Server(*name)
	if err := server.Load(); err != nil {
		return fmt.Errorf("failed to load server: %w", err)
	}

	if err := server.Destroy(*force); err != nil {
		return fmt.Errorf("failed to destroy server: %w", err)
	}

	if *purge {
		time.Sleep(15 * time.Second)
		if err := server.Purge(*force); err != nil {
			return fmt.Errorf("failed to purge server: %w", err)
		}
	}

	fmt.Printf("Server %s destroyed successfully.\n", *name)
	return nil
}
