package congo_auth

import (
	"embed"
	"log"
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
)

type CongoAuth struct {
	DB             *congo.Database
	CookieName     string
	SetupView      string
	SetupRedirect  string
	SignupCallback func(*Controller, *Identity) http.HandlerFunc
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
		SetupView:      "congo-signup.html",
		SigninViews:    map[string]string{},
		LogoutRedirect: "/",
	}
	if err := dir.DB.MigrateUp(); err != nil {
		log.Fatal("Failed to setup auth database:", err)
	}
	for _, opt := range opts {
		opt(dir)
	}
	return dir
}

type DirectoryOpt func(*CongoAuth)

func WithCookieName(name string) DirectoryOpt {
	return func(d *CongoAuth) {
		if name == "" {
			log.Fatal("cannot have empty cookie name")
		}
		d.CookieName = name
	}
}

func WithSetupView(view, dest string) DirectoryOpt {
	return func(auth *CongoAuth) {
		auth.SetupView = view
		auth.SetupRedirect = dest
	}
}

func WithSignupCallback(fn func(*Controller, *Identity) http.HandlerFunc) DirectoryOpt {
	return func(auth *CongoAuth) {
		auth.SignupCallback = fn
	}
}

func WithSigninView(role, view string) DirectoryOpt {
	return func(auth *CongoAuth) {
		auth.SigninViews[role] = view
	}
}

func WithSigninDest(url string) DirectoryOpt {
	return func(auth *CongoAuth) {
		auth.SigninRedirect = url
	}
}

func WithLogoutRedirect(url string) DirectoryOpt {
	return func(d *CongoAuth) {
		if url == "" {
			log.Fatal("cannot have empty logout redirect url")
		}
		d.LogoutRedirect = url
	}
}
