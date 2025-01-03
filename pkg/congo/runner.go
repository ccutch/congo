package congo

import (
	"cmp"
	"log"
	"os"
)

type Service interface {
	Start() error
}

func (app *Application) StartFromEnv(services ...Service) {
	app.WithCredentials(envCredentials())
	Start(app, services...)
}

func envCredentials() (string, string) {
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

func Start(main Service, services ...Service) {
	for i, s := range services {
		go func() {
			if err := s.Start(); err != nil {
				log.Fatalf("Service %d failed: %s", i, err)
			}
		}()
	}

	if err := main.Start(); err != nil {
		log.Fatalf("Failed to start main service: %s", err)
	}

	log.Fatal("Main service stopped without error")
}

type Ignorer struct {
	service Service
}

func IgnoreError(service Service) *Ignorer {
	return &Ignorer{service}
}

func (i *Ignorer) Start() error {
	err := i.service.Start()
	if err != nil {
		log.Printf("Failed to run service: %s", err)
	}

	return nil
}
