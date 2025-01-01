package congo_boot

import "log"

type Ignorer struct {
	name    string
	service Service
}

func IgnoreErr(name string, service Service) *Ignorer {
	return &Ignorer{name, service}
}

func (i *Ignorer) Start() error {
	log.Println("Starting service: ", i.name)
	err := i.service.Start()
	if err != nil {
		log.Printf("Failed to run service: %s", err)
	}
	return nil
}
