package congo_code

import (
	"embed"
	"log"

	"github.com/ccutch/congo/pkg/congo"
)

//go:embed all:migrations
var migrations embed.FS

type CongoCode struct {
	db *congo.Database
}

func InitCongoCode(root string) *CongoCode {
	code := CongoCode{
		db: congo.SetupDatabase(root, "code.db", migrations),
	}
	if err := code.db.MigrateUp(); err != nil {
		log.Fatal("Failed to setup code db:", err)
	}
	return &code
}
