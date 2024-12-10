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
	)

	if err := cmd.Parse(args[1:]); err != nil {
		return err
	}

	if *apiKey == "$DIGITAL_OCEAN_API_KEY" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	client := congo_host.NewClient(*path, *apiKey)
	server, err := client.LoadServer(*name, *region)
	if err != nil {
		return err
	}

	if *rebuild && *binary == "" {
		exec.Command("go", "build", "-o", "congo", ".").Run()
		*binary = "./congo"
	}

	if *binary != "" {
		*binary, server.Err = filepath.Abs(*binary)
		server.Err = server.Copy(*binary, "/root/congo")
	}

	server.Start()
	return server.Err
}
