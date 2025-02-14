package main

import (
	"cmp"
	"flag"
	"os"

	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_host/platforms/digitalocean"
	"github.com/pkg/errors"
)

func genCerts(args ...string) error {
	var (
		cmd        = flag.NewFlagSet("gen-certs", flag.ExitOnError)
		apiKey     = cmd.String("api-key", "$DIGITAL_OCEAN_API_KEY", "Digital Ocean API Key")
		path       = cmd.String("data-path", "/tmp/congo", "Local storage for SSH Keys")
		name       = cmd.String("name", "congo-server", "Name of Digital Ocean droplet")
		domainName = cmd.String("domain", "", "Domain name to generate cert for")
		admin      = cmd.String("admin", "", "admin email")
		assign     = cmd.Bool("assign", false, "Assign domain to server")
	)

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
		return errors.Wrap(err, "failed to get server")
	}

	if err := server.Reload(); err != nil {
		return errors.Wrap(err, "failed to reload server")
	}

	*admin = cmp.Or(*admin, "admin@testing.com")
	if *domainName != "" {
		domain := server.Domain(*domainName)
		if err = domain.Save(); err != nil {
			return errors.Wrap(err, "failed to save domain")
		}

		if *assign {
			if err := server.Assign(domain); err != nil {
				return errors.Wrap(err, "failed to assign domain")
			}
		}

		if err := domain.Verify(*admin); err != nil {
			return errors.Wrap(err, "failed to verify domain")
		}
	} else {
		domains, err := server.Domains()
		if err != nil {
			return errors.Wrap(err, "failed to get domains")
		}

		otherDomains := []*congo_host.Domain{}
		for _, d := range domains {
			if d.Verified {
				otherDomains = append(otherDomains, d)
			}
		}

		if err := server.Verify(*admin, otherDomains...); err != nil {
			return errors.Wrap(err, "failed to verify domain")
		}
	}

	return server.Restart()
}
