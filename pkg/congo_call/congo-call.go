package congo_call

import (
	"embed"
	"log"

	"github.com/ccutch/congo/pkg/congo"
)

//go:embed all:migrations
var migrations embed.FS

type CongoCall struct {
	db *congo.Database
}

func InitCongoCall(root string) *CongoCall {
	db := congo.SetupDatabase(root, "call.db", migrations)
	if err := db.MigrateUp(); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
	return &CongoCall{db: db}
}
