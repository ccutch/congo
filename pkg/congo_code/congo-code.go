package congo_code

import (
	"embed"
	"log"

	"github.com/ccutch/congo/pkg/congo"
)

//go:embed all:migrations
var migrations embed.FS

type CongoCode struct {
	DB *congo.Database
}

func InitCongoCode(root string, opts ...CongoCodeOpt) *CongoCode {
	code := CongoCode{DB: congo.SetupDatabase(root, "code.db", migrations)}
	if err := code.DB.MigrateUp(); err != nil {
		log.Fatal("Failed to setup code db:", err)
	}

	for _, opt := range opts {
		if err := opt(&code); err != nil {
			log.Fatal("Failed to setup Congo Code: ", err)
		}
	}

	return &code
}

type CongoCodeOpt func(*CongoCode) error
