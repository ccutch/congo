package main

import (
	"flag"
	"os"

	"github.com/ccutch/congo/pkg/congo_host"
)

func connect(args ...string) error {
	var (
		cmd    = flag.NewFlagSet("connect", flag.ExitOnError)
		apiKey = cmd.String("api-key", "$DIGITAL_OCEAN_API_KEY", "Digital Ocean API Key")
		path   = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		name   = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
	)

	if err := cmd.Parse(args[1:]); err != nil {
		return err
	}

	if *apiKey == "$DIGITAL_OCEAN_API_KEY" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	host := congo_host.InitCongoHost(*path, congo_host.WithApiToken(*apiKey))
	server, err := host.LoadServer(*name)
	if err != nil {
		return err
	}

	return server.Run(cmd.Args()...)
}
