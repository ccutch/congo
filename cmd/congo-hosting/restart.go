package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_host/platforms/digitalocean"
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
		web    = cmd.String("web", "", "Website we are deploying")
		enVars stringArray
	)

	cmd.Var(&enVars, "env", "Environment variables to include in environment")
	if err := cmd.Parse(args[1:]); err != nil {
		return err
	}

	if *apiKey == "$DIGITAL_OCEAN_API_KEY" {
		*apiKey = os.Getenv("DIGITAL_OCEAN_API_KEY")
	}

	host := congo_host.InitCongoHost(*path,
		congo_host.WithPlatform(digitalocean.NewClient(*apiKey)))
	server, err := host.GetServer(*name)
	if err != nil {
		return err
	}

	if err := server.Reload(); err != nil {
		return err
	}

	if *app != "" {
		log.Println("Building binary...")
		if err = exec.Command("go", "build", "-o", "congo", "./apps/"+*app).Run(); err != nil {
			log.Println("Failed to build binary: ", err)
			return errors.Wrap(err, "failed to build binary")
		}
		*binary = "./congo"
	} else if *web != "" {
		log.Println("Building website...")
		if err = exec.Command("go", "build", "-o", "congo", "./web/"+*web).Run(); err != nil {
			log.Println("Failed to build website: ", err)
			return errors.Wrap(err, "failed to build website")
		}
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
