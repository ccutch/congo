package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

type AuthController struct {
	*congo_auth.Controller
}

func (auth *AuthController) Setup(app *congo.Application) {
	auth.Controller.CongoAuth = congo_auth.InitCongoAuth(app.DB.Root,
		congo_auth.WithCookieName("workhouse"),
		congo_auth.WithSetupView("setup.html", "/"),
		congo_auth.WithAccessView("user", "welcome.html"),
		congo_auth.WithAccessView("developer", "signin.html"))

	auth.Controller.Setup(app)
	app.HandleFunc("POST /_auth/signin", auth.customSignin)
}

func (auth AuthController) Handle(req *http.Request) congo.Controller {
	return &auth
}

func (auth *AuthController) Developers() ([]*congo_auth.Identity, error) {
	return auth.CongoAuth.SearchByRole("developers", auth.URL.Query().Get("query"))
}

func (auth *AuthController) Users() ([]*congo_auth.Identity, error) {
	return auth.CongoAuth.SearchByRole("users", auth.URL.Query().Get("query"))
}

func (auth AuthController) customSignin(w http.ResponseWriter, r *http.Request) {
	i, err := auth.CongoAuth.Lookup(r.FormValue("username"))
	if err != nil {
		auth.Render(w, r, "error-message", fmt.Errorf("failed to find identity"))
		return
	}

	if !i.Verify(r.FormValue("password")) {
		auth.Render(w, r, "error-message", fmt.Errorf("failed to find identity"))
		return
	}

	s, err := i.NewSession()
	if err != nil {
		auth.Render(w, r, "error-message", fmt.Errorf("failed to start session: %s", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.CongoAuth.CookieName + "-" + i.Role,
		Path:     "/",
		Value:    s.Token(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	if i.Role == "user" {
		auth.Redirect(w, r, "/")
		return
	}

	if i.Role == "developer" {
		auth.Redirect(w, r, "/posts")
		return
	}

	auth.Refresh(w, r)
}
