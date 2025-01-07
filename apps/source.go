package apps

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed all:*/*.go all:*/**/*.go
var SourceFiles embed.FS

//go:embed all:*/templates/*.html all:*/templates/**/*.html all:*/migrations/*.sql
var ResourceFiles embed.FS

func Apps() []string {
	apps := []string{}
	fs.WalkDir(SourceFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}
		apps = append(apps, d.Name())
		return nil
	})
	log.Println("Apps: ", apps)
	return apps
}
