package congo_auth

import (
	"embed"
	"errors"
	"log"

	"github.com/ccutch/congo/pkg/congo"
)

type CongoAuth struct {
	DB             *congo.Database
	CookieName     string
	LogoutRedirect string
	SetupView      string
	LoginView      string
	DefaultRole    string
	defaultRoles   []string
}

//go:embed all:migrations
var migrations embed.FS

func InitCongoAuth(root string, opts ...DirectoryOpt) *CongoAuth {
	dir := &CongoAuth{
		DB:             congo.SetupDatabase(root, "auth.db", migrations),
		CookieName:     "congo-app",
		DefaultRole:    "user",
		LogoutRedirect: "/",
		LoginView:      "congo-signin.html",
		defaultRoles:   []string{"user"},
	}
	if err := dir.DB.MigrateUp(); err != nil {
		log.Fatal("Failed to setup auth database:", err)
	}
	for _, opt := range opts {
		if err := opt(dir); err != nil {
			log.Fatalf("Failed to open Directory @ %s: %s", root, err)
		}
	}
	return dir
}

type DirectoryOpt func(*CongoAuth) error

func WithCookieName(name string) DirectoryOpt {
	return func(d *CongoAuth) error {
		if name == "" {
			return errors.New("cannot have empty cookie name")
		}
		d.CookieName = name
		return nil
	}
}

func WithDefaultRole(role string) DirectoryOpt {
	return func(d *CongoAuth) error {
		if role == "" {
			return errors.New("cannot have empty default role")
		}
		if len(d.defaultRoles) == 1 && d.defaultRoles[0] == d.DefaultRole {
			d.defaultRoles = []string{role}
		}
		d.DefaultRole = role
		return nil
	}
}

func WithDefaultRoles(roles ...string) DirectoryOpt {
	return func(d *CongoAuth) error {
		if len(roles) == 0 {
			roles = append(roles, d.DefaultRole)
		}
		d.defaultRoles = roles
		return nil
	}
}

func WithLogoutRedirect(url string) DirectoryOpt {
	return func(d *CongoAuth) error {
		if url == "" {
			return errors.New("cannot have empty logout redirect url")
		}
		d.LogoutRedirect = url
		return nil
	}
}

func WithSetupView(view string) DirectoryOpt {
	return func(auth *CongoAuth) error {
		auth.SetupView = view
		return nil
	}
}

func WithLoginView(view string) DirectoryOpt {
	return func(auth *CongoAuth) error {
		auth.LoginView = view
		return nil
	}
}
