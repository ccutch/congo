package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ccutch/congo/pkg/congo_host"
)

func restart(args ...string) error {
	var (
		cmd     = flag.NewFlagSet("restart", flag.ExitOnError)
		apiKey  = cmd.String("api-key", "$DIGITAL_OCEAN_API_KEY", "Digital Ocean API Key")
		path    = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		name    = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		region  = cmd.String("region", "sfo2", "Region of Digital Ocean droplet")
		rebuild = cmd.Bool("rebuild", false, "Rebuld local directory as Congo binary")
		binary  = cmd.String("binary", "", "Local binary to copy to Digital Ocean droplet")
		app     = cmd.String("app", "", "Prototype to use for the server")
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
		return err
	}

	if (*app != "" || *rebuild) && *binary == "" {
		exec.Command("go", "build", "-o", "congo", "./apps/"+*app).Run()
		*binary = "./congo"
	}

	if *binary != "" {
		*binary, server.Error = filepath.Abs(*binary)
		server.Error = server.Copy(*binary, "/root/congo")
	}

	server.Start()
	return server.Error
}
