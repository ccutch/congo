package congo_boot

import (
	"cmp"
	"log"
	"os"

	"github.com/ccutch/congo/pkg/congo"
)

type Service interface {
	Start() error
}

func StartFromEnv(app *congo.Application, services ...Service) {
	app.WithCredentials(LoadEnv(app))
	Start(app, services...)
}

func LoadEnv(app *congo.Application) (string, string) {
	cert := os.Getenv("CONGO_SSL_FULLCHAIN")
	cert = cmp.Or(cert, "/root/fullchain.pem")
	if _, err := os.Stat(cert); err != nil {
		return "", ""
	}

	key := os.Getenv("CONGO_SSL_PRIVKEY")
	key = cmp.Or(key, "/root/privkey.pem")
	if _, err := os.Stat(key); err != nil {
		return "", ""
	}

	return cert, key
}

func Start(app *congo.Application, services ...Service) {
	for i, s := range services {
		go func() {
			if err := s.Start(); err != nil {
				log.Fatalf("Service %d failed: %s", i, err)
			}
		}()
	}

	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start main service: %s", err)
	}

	log.Fatal("Main service stopped without error")
}
