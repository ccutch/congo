package congo_auth

import (
	"embed"
	"errors"
	"log"

	"github.com/ccutch/congo/pkg/congo"
)

type Directory struct {
	DB             *congo.Database
	CookieName     string
	DefaultRole    string
	LogoutRedirect string
}

//go:embed all:migrations
var migrations embed.FS

func OpenDirectory(app *congo.Application, opts ...DirectoryOpt) *Directory {
	dir := &Directory{DB: congo.SetupDatabase(app.DB.Root, "directory.sql", migrations)}
	for _, opt := range opts {
		if err := opt(dir); err != nil {
			log.Fatalf("Failed to open Directory @ %s: %s", app.DB.Root, err)
		}
	}
	return dir
}

type DirectoryOpt func(*Directory) error

func WithCookieName(name string) DirectoryOpt {
	return func(d *Directory) error {
		if name == "" {
			return errors.New("cannot have empty cookie name")
		}
		d.CookieName = name
		return nil
	}
}

func WithDefaultRole(role string) DirectoryOpt {
	return func(d *Directory) error {
		if role == "" {
			return errors.New("cannot have empty default role")
		}
		d.DefaultRole = role
		return nil
	}
}
func WithLogoutRedirect(url string) DirectoryOpt {
	return func(d *Directory) error {
		if url == "" {
			return errors.New("cannot have empty logout redirect url")
		}
		d.LogoutRedirect = url
		return nil
	}
}
