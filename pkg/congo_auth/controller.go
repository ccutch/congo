package congo_auth

import (
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
	*CongoAuth
}

func (auth *CongoAuth) Controller() (string, *Controller) {
	return "auth", &Controller{CongoAuth: auth}
}

func (auth *Controller) Setup(app *congo.Application) {
	auth.BaseController.Setup(app)
	app.WithTemplates(Templates)
	app.HandleFunc("POST /_auth/signup/{role}", auth.handleSignup)
	app.HandleFunc("POST /_auth/signin/{role}", auth.handleSignin)
	app.HandleFunc("POST /_auth/logout/{role}", auth.handleLogout)
	app.HandleFunc("DELETE /_auth/session/{id}", auth.endSession)
}

func (auth Controller) Handle(r *http.Request) congo.Controller {
	auth.Request = r
	return &auth
}

func (auth *Controller) Current(role string) *Identity {
	identity, _ := auth.CongoAuth.Authenticate(role, auth.Request)
	return identity
}

func (auth *Controller) Usage() ([]*Usage, error) {
	identity := auth.Current(auth.PathValue("role"))
	return identity.Usages()
}

func (auth *Controller) Identities() ([]*Identity, error) {
	role := auth.PathValue("role")
	if role != "" {
		return auth.SearchByRole(role, auth.URL.Query().Get("query"))
	}

	var identities []*Identity
	imap, err := auth.Search(auth.URL.Query().Get("query"))
	for _, idents := range imap {
		identities = append(identities, idents...)
	}

	return identities, err
}

func (auth Controller) handleSignup(w http.ResponseWriter, r *http.Request) {
	email, username, password := r.FormValue("email"), r.FormValue("username"), r.FormValue("password")
	if email == "" || username == "" || password == "" {
		auth.Render(w, r, "error-message", fmt.Errorf("missing required fields"))
		return
	}

	identity, err := auth.CongoAuth.Create(r.PathValue("role"), email, username, password)
	if err != nil {
		auth.Render(w, r, "error-message", fmt.Errorf("failed to create identity: %s", err))
		return
	}

	session, err := identity.NewSession()
	if err != nil {
		auth.Render(w, r, "error-message", fmt.Errorf("failed to start session: %s", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.CongoAuth.CookieName + "-" + r.PathValue("role"),
		Path:     "/",
		Value:    session.Token(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	if auth.CongoAuth.SignupCallback != nil {
		auth.CongoAuth.SignupCallback(&auth, identity).ServeHTTP(w, r)
	} else {
		if auth.CongoAuth.SetupRedirect != "" {
			http.Redirect(w, r, auth.CongoAuth.SetupRedirect, http.StatusFound)
			return
		}
		auth.Refresh(w, r)
	}
}

func (auth Controller) handleSignin(w http.ResponseWriter, r *http.Request) {
	identity, err := auth.CongoAuth.Lookup(r.FormValue("username"))
	if err != nil {
		auth.Render(w, r, "error-message", fmt.Errorf("failed to find identity"))
		return
	}

	if !identity.Verify(r.FormValue("password")) {
		auth.Render(w, r, "error-message", fmt.Errorf("failed to find identity"))
		return
	}

	session, err := identity.NewSession()
	if err != nil {
		auth.Render(w, r, "error-message", fmt.Errorf("failed to start session: %s", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.CongoAuth.CookieName + "-" + r.PathValue("role"),
		Path:     "/",
		Value:    session.Token(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	if auth.CongoAuth.SigninRedirect != "" {
		auth.Redirect(w, r, auth.CongoAuth.SigninRedirect)
		return
	}

	auth.Refresh(w, r)
}

func (auth Controller) handleLogout(w http.ResponseWriter, r *http.Request) {
	role := r.PathValue("role")
	if _, s := auth.CongoAuth.Authenticate(role, r); s != nil {
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

	auth.Redirect(w, r, auth.CongoAuth.LogoutRedirect)
}

func (auth Controller) endSession(w http.ResponseWriter, r *http.Request) {
	session, err := auth.CongoAuth.GetSession(r.PathValue("id"))
	if err != nil {
		auth.Render(w, r, "error-message", err)
		return
	}

	if err = session.Delete(); err != nil {
		auth.Render(w, r, "error-message", err)
		return
	}

	auth.Redirect(w, r, "/")
}
