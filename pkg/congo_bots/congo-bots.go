package congo_bots

import (
	"fmt"
	"log"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_host"
)

type CongoBots struct {
	DB     *congo.Database
	client *congo_host.Service
	models map[string]bool
}

func InitCongoBots(host congo_host.CongoHost, opts ...CongoBotsOpt) *CongoBots {
	db := congo.SetupDatabase(host.DB.Root, "bots.db", nil)
	if err := db.MigrateUp(); err != nil {
		log.Fatal("Failed to setup bots db:", err)
	}
	service := host.Local().Service("ollama",
		congo_host.WithImage("ollama/ollama"),
		congo_host.WithTag("latest"),
		congo_host.WithVolume(fmt.Sprintf("%s/services/ollama:/root/.ollama", host.DB.Root)))
	go service.Start()
	return &CongoBots{db, service, map[string]bool{}}
}

type CongoBotsOpt func(*CongoBots)

func WithModel(name string) CongoBotsOpt {
	return func(bots *CongoBots) {
		bots.models[name] = false
		go func() {
			// client, err := api.ClientFromEnvironment()
			// if err != nil {
			// 	log.Fatal("Failed to create ollama client:", err)
			// }

			// err = client.Pull(context.Background(), &api.PullRequest{Model: name}, nil)
			// if err != nil {
			// 	log.Fatal("Failed to pull ollama model:", err)
			// }

			// bots.models[name] = true
		}()
	}
}
