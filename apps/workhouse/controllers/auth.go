package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

type AuthController struct {
	*congo_auth.AuthController
}

func NewAuthController(auth *congo_auth.CongoAuth) *AuthController {
	return &AuthController{&congo_auth.AuthController{CongoAuth: auth}}
}

func (auth *AuthController) Setup(app *congo.Application) {
	auth.BaseController.Setup(app)
	app.HandleFunc("POST /_auth/signup", auth.handleSignup)
	app.HandleFunc("POST /_auth/signin", auth.handleSignin)
	app.HandleFunc("POST /_auth/signout", auth.handleSignout)
}

func (auth AuthController) Handle(req *http.Request) congo.Controller {
	auth.Request = req
	return &auth
}

func (auth *AuthController) SelectedUser() *congo_auth.Identity {
	i, err := auth.CongoAuth.Lookup(auth.PathValue("user"))
	if err != nil {
		return nil
	}
	return i
}

func (auth *AuthController) CurrentUser() *congo_auth.Identity {
	i, _ := auth.CongoAuth.Authenticate(auth.Request, "user", "developer")
	return i
}

func (auth *AuthController) Developers() ([]*congo_auth.Identity, error) {
	return auth.CongoAuth.SearchByRole("developer", auth.URL.Query().Get("query"))
}

func (auth *AuthController) Users() ([]*congo_auth.Identity, error) {
	return auth.CongoAuth.SearchByRole("user", auth.URL.Query().Get("query"))
}

func (auth AuthController) handleSignup(w http.ResponseWriter, r *http.Request) {
	email, username, password := r.FormValue("email"), r.FormValue("username"), r.FormValue("password")
	if email == "" || username == "" || password == "" {
		auth.Render(w, r, "error-message", fmt.Errorf("missing required fields"))
		return
	}

	role := "user"
	if auth.CongoAuth.Count() == 0 {
		role = "developer"
	}

	i, err := auth.CongoAuth.Create(role, email, username, password)
	if err != nil {
		auth.Render(w, r, "error-message", fmt.Errorf("failed to create identity: %s", err))
		return
	}

	s, err := i.NewSession()
	if err != nil {
		auth.Render(w, r, "error-message", fmt.Errorf("failed to start session: %s", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.CongoAuth.CookieName + "-" + r.PathValue("role"),
		Path:     "/",
		Value:    s.Token(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	if i.Role == "user" {
		auth.Redirect(w, r, "/")
		return
	}

	go func() {
		devs, _ := auth.Developers()
		c := auth.Use("content").(*ContentController)
		c.Code.RunWorkspace(c.Host, i.Name+"-workspace", 7000+len(devs), c.Repo)
	}()

	auth.Redirect(w, r, "/code")
}

func (auth AuthController) handleSignin(w http.ResponseWriter, r *http.Request) {
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
		auth.Redirect(w, r, "/code")
		return
	}

	auth.Refresh(w, r)
}

func (auth AuthController) handleSignout(w http.ResponseWriter, r *http.Request) {
	for role := range auth.CongoAuth.SigninViews {
		if _, s := auth.CongoAuth.Authenticate(r, role); s != nil {
			if err := s.Delete(); err != nil {
				auth.Render(w, r, "error-message", err)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     auth.CongoAuth.CookieName + "-" + role,
				Path:     "/",
				Value:    "",
				Expires:  time.Now().Add(-1 * time.Hour),
				HttpOnly: true,
			})
		}
	}

	auth.Redirect(w, r, auth.CongoAuth.LogoutRedirect)
}
