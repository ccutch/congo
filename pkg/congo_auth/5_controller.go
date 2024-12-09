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

func (ctrl *Controller) OnMount(app *congo.Application) error {
	ctrl.Application = app
	app.HandleFunc("POST /_auth/signup", ctrl.handleSignup)
	app.HandleFunc("POST /_auth/signin", ctrl.handleSignin)
	app.HandleFunc("POST /_auth/usage", ctrl.handleMyUsage)
	app.HandleFunc("POST /_auth/logout", ctrl.handleLogout)
	return nil
}

func (ctrl Controller) OnRequest(r *http.Request) congo.Controller {
	ctrl.Request = r
	return &ctrl
}

func (ctrl Controller) handleSignup(app *congo.Application, w http.ResponseWriter, r *http.Request) {
	role := cmp.Or(r.FormValue("role"), ctrl.dir.DefaultRole)
	email, username, password := r.FormValue("email"), r.FormValue("username"), r.FormValue("password")
	if email == "" || username == "" || password == "" {
		ctrl.Render(app, w, r, "signup-form", fmt.Errorf("missing required fields"))
		return
	}
	identity, err := ctrl.dir.Create(role, email, username, password)
	if err != nil {
		ctrl.Render(app, w, r, "signup-form", fmt.Errorf("failed to create identity: %s", err))
		return
	}
	session, err := identity.NewSession()
	if err != nil {
		ctrl.Render(app, w, r, "signup-form", fmt.Errorf("failed to start session: %s", err))
		return
	}
	r.AddCookie(&http.Cookie{
		Name:     ctrl.dir.CookieName,
		Path:     "/",
		Value:    session.Token(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
	ctrl.Refresh(w, r)
}

func (ctrl Controller) handleSignin(app *congo.Application, w http.ResponseWriter, r *http.Request) {
	identity, err := ctrl.dir.Lookup(r.FormValue("username"))
	if err != nil {
		ctrl.Render(app, w, r, "signin-form", fmt.Errorf("failed to find identity"))
		return
	}
	if !identity.Verify(r.FormValue("password")) {
		ctrl.Render(app, w, r, "signin-form", fmt.Errorf("failed to find identity"))
		return
	}
	session, err := identity.NewSession()
	if err != nil {
		ctrl.Render(app, w, r, "signin-form", fmt.Errorf("failed to start session: %s", err))
		return
	}
	r.AddCookie(&http.Cookie{
		Name:     ctrl.dir.CookieName,
		Path:     "/",
		Value:    session.Token(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
	ctrl.Refresh(w, r)
}

func (ctrl Controller) handleMyUsage(app *congo.Application, w http.ResponseWriter, r *http.Request) {
	i, _ := ctrl.dir.Authenticate(r)
	if i == nil {
		ctrl.Render(app, w, r, "signin-form", nil)
		return
	}
	usages, err := i.Usages()
	ctrl.Render(app, w, r, "my-usages", struct {
		Usages []*Usage
		Error  error
	}{usages, err})
}

func (ctrl Controller) handleLogout(app *congo.Application, w http.ResponseWriter, r *http.Request) {
	if _, s := ctrl.dir.Authenticate(r); s != nil {
		s.End()
		r.AddCookie(&http.Cookie{
			Name:     ctrl.dir.CookieName,
			Path:     "/",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour),
			HttpOnly: true,
		})
	}
	ctrl.Redirect(w, r, ctrl.dir.LogoutRedirect)
}
