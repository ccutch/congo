package congo_auth

import (
	"embed"
	"errors"
	"log"

	"github.com/ccutch/congo/pkg/congo"
)

type Directory struct {
	app            *congo.Application
	DB             *congo.Database
	CookieName     string
	DefaultRole    string
	LogoutRedirect string
}

//go:embed all:migrations
var migrations embed.FS

func InitCongoAuth(app *congo.Application, opts ...DirectoryOpt) *Directory {
	dir := &Directory{
		app:            app,
		DB:             congo.SetupDatabase(app.DB.Root, "directory.db", migrations),
		CookieName:     "congo-app",
		DefaultRole:    "user",
		LogoutRedirect: "/",
	}
	if err := dir.DB.MigrateUp(); err != nil {
		log.Fatal("Failed to setup auth database:", err)
	}
	for _, opt := range opts {
		if err := opt(dir); err != nil {
			log.Fatalf("Failed to open Directory @ %s: %s", app.DB.Root, err)
		}
	}
	app.WithController("auth", dir.Controller())
	app.WithTemplates(Templates)
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
