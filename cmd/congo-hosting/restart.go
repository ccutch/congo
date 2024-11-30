package main

import (
	"flag"
	"os"

	"congo.gitpost.app/internal/hosting"
)

func restart(args ...string) error {
	var (
		cmd    = flag.NewFlagSet("restart", flag.ExitOnError)
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
		return err
	}

	server.Start()
	return server.Err
}