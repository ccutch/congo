package congo_boot

import "log"

type Ignorer struct {
	service Service
}

func IgnoreErr(service Service) *Ignorer {
	return &Ignorer{service}
}

func (i *Ignorer) Start() error {
	err := i.service.Start()
	if err != nil {
		log.Printf("Failed to run service: %s", err)
	}
	return nil
}
