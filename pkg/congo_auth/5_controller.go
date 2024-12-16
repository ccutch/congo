package congo_auth

import (
	"cmp"
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/ccutch/congo/pkg/congo"
)

//go:embed all:templates
var Templates embed.FS

type Controller struct {
	congo.BaseController
	dir *Directory
}

func (dir *Directory) Controller() *Controller {
	return &Controller{congo.BaseController{}, dir}
}

func (auth *Controller) Setup(app *congo.Application) {
	auth.Application = app
	app.HandleFunc("GET /_auth/signup/{role}", auth.Template("congo-auth/signup-form"))
	app.HandleFunc("POST /_auth/signup/{role}", auth.handleSignup)
	app.HandleFunc("GET /_auth/signin/{role}", auth.Template("congo-auth/signin-form"))
	app.HandleFunc("POST /_auth/signin/{role}", auth.handleSignin)
	app.HandleFunc("GET /_auth/usage/{role}", auth.handleMyUsage)
	app.HandleFunc("GET /_auth/logout/{role}", auth.handleLogout)
}

func (auth Controller) Handle(r *http.Request) congo.Controller {
	auth.Request = r
	return &auth
}

func (auth *Controller) Current(role string) *Identity {
	identity, _ := auth.dir.Authenticate(role, auth.Request)
	return identity
}

func (auth Controller) handleSignup(w http.ResponseWriter, r *http.Request) {
	role := cmp.Or(r.FormValue("role"), auth.dir.DefaultRole)
	email, username, password := r.FormValue("email"), r.FormValue("username"), r.FormValue("password")
	if email == "" || username == "" || password == "" {
		auth.Render(w, r, "congo-auth/error-message", fmt.Errorf("missing required fields"))
		return
	}
	identity, err := auth.dir.Create(role, email, username, password)
	if err != nil {
		auth.Render(w, r, "congo-auth/error-message", fmt.Errorf("failed to create identity: %s", err))
		return
	}
	session, err := identity.NewSession()
	if err != nil {
		auth.Render(w, r, "congo-auth/error-message", fmt.Errorf("failed to start session: %s", err))
		return
	}
	token := session.Token()
	http.SetCookie(w, &http.Cookie{
		Name:     auth.dir.CookieName + "-" + role,
		Path:     "/",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
	auth.Refresh(w, r)
}

func (auth Controller) handleSignin(w http.ResponseWriter, r *http.Request) {
	identity, err := auth.dir.Lookup(r.FormValue("username"))
	if err != nil {
		auth.Render(w, r, "congo-auth/error-message", fmt.Errorf("failed to find identity"))
		return
	}
	if !identity.Verify(r.FormValue("password")) {
		auth.Render(w, r, "congo-auth/error-message", fmt.Errorf("failed to find identity"))
		return
	}
	session, err := identity.NewSession()
	if err != nil {
		auth.Render(w, r, "congo-auth/error-message", fmt.Errorf("failed to start session: %s", err))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.dir.CookieName + "-" + r.PathValue("role"),
		Path:     "/",
		Value:    session.Token(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
	auth.Refresh(w, r)
}

func (auth Controller) handleMyUsage(w http.ResponseWriter, r *http.Request) {
	role := r.PathValue("role")
	i, _ := auth.dir.Authenticate(role, r)
	if i == nil {
		auth.Render(w, r, "congo-auth/signin-form", role)
		return
	}
	usages, err := i.Usages()
	auth.Render(w, r, "my-usages", struct {
		Usages []*Usage
		Error  error
	}{usages, err})
}

func (auth Controller) handleLogout(w http.ResponseWriter, r *http.Request) {
	role := r.PathValue("role")
	if _, s := auth.dir.Authenticate(role, r); s != nil {
		s.End()
		http.SetCookie(w, &http.Cookie{
			Name:     auth.dir.CookieName + "-" + role,
			Path:     "/",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour),
			HttpOnly: true,
		})
	}
	auth.Redirect(w, r, auth.dir.LogoutRedirect)
}
