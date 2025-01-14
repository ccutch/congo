package congo_book

import (
	"log"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

type CongoBook struct {
	db   *congo.Database
	auth *congo_auth.CongoAuth
}

func InitCongoBook(auth *congo_auth.CongoAuth) *CongoBook {
	db := congo.SetupDatabase(auth.DB.Root, "book.db", nil)
	if err := db.MigrateUp(); err != nil {
		log.Fatal("Failed to setup book db:", err)
	}
	return &CongoBook{db, auth}
}
