package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_host/backend/digitalocean"
	"github.com/pkg/errors"
)

func restart(args ...string) error {
	var (
		cmd    = flag.NewFlagSet("restart", flag.ExitOnError)
		apiKey = cmd.String("api-key", "$DIGITAL_OCEAN_API_KEY", "Digital Ocean API Key")
		path   = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		name   = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		binary = cmd.String("binary", "", "Local binary to copy to Digital Ocean droplet")
		app    = cmd.String("app", "", "Prototype to use for the server")
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
		return err
	}

	if err := server.Reload(); err != nil {
		return err
	}

	if *app != "" && *binary == "" {
		exec.Command("go", "build", "-o", "congo", "./apps/"+*app).Run()
		*binary = "./congo"
	}

	if *binary != "" {
		if *binary, err = filepath.Abs(*binary); err != nil {
			return errors.Wrap(err, "failed to get absolute path")
		}
		if _, _, err = server.Copy(*binary, "/root/congo"); err != nil {
			return errors.Wrap(err, "failed to copy binary to server")
		}
	}

	return server.Restart()
}
