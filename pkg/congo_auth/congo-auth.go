package congo_auth

import (
	"cmp"
	"embed"
	"errors"
	"log"

	"github.com/ccutch/congo/pkg/congo"
)

type CongoAuth struct {
	DB             *congo.Database
	CookieName     string
	DefaultRole    string
	SetupView      string
	SetupRedirect  string
	SigninViews    map[string]string
	SigninRedirect string
	LogoutRedirect string
}

//go:embed all:migrations
var migrations embed.FS

func InitCongoAuth(root string, opts ...DirectoryOpt) *CongoAuth {
	dir := &CongoAuth{
		DB:             congo.SetupDatabase(root, "auth.db", migrations),
		CookieName:     "congo-app",
		DefaultRole:    "user",
		SetupView:      "congo-signup.html",
		SigninViews:    map[string]string{"user": "congo-signin.html"},
		LogoutRedirect: "/",
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
		d.DefaultRole = role
		d.SigninViews[role] = cmp.Or(d.SigninViews[role], "congo-signin.html")
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

func WithSigninDest(url string) DirectoryOpt {
	return func(auth *CongoAuth) error {
		auth.SigninRedirect = url
		return nil
	}
}
func WithSetupView(view string) DirectoryOpt {
	return func(auth *CongoAuth) error {
		auth.SetupView = view
		return nil
	}
}

func WithSigninView(role, view string) DirectoryOpt {
	return func(auth *CongoAuth) error {
		auth.SigninViews[role] = view
		return nil
	}
}

func WithSetupDest(url string) DirectoryOpt {
	return func(auth *CongoAuth) error {
		auth.SetupRedirect = url
		return nil
	}
}
