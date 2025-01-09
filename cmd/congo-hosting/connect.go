package main

import (
	"flag"
	"os"

	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_host/backend/digitalocean"
	"github.com/pkg/errors"
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

	host := congo_host.InitCongoHost(*path, digitalocean.NewClient(*apiKey))
	server, err := host.GetServer(*name)
	if err != nil {
		return errors.Wrap(err, "failed to get server")
	}

	if err := server.Reload(); err != nil {
		return err
	}

	_, out, err := server.Run(nil, cmd.Args()...)
	return errors.Wrap(err, out.String())
}
