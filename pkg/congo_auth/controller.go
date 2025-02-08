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

type AuthController struct {
	congo.BaseController
	*CongoAuth
}

func (auth *CongoAuth) Controller() (string, *AuthController) {
	return "auth", &AuthController{CongoAuth: auth}
}

func (auth *AuthController) Setup(app *congo.Application) {
	auth.BaseController.Setup(app)
	app.WithTemplates(Templates)
	http.HandleFunc("POST /_auth/signup/{role}", auth.handleSignup)
	http.HandleFunc("POST /_auth/signin/{role}", auth.handleSignin)
	http.HandleFunc("POST /_auth/logout/{role}", auth.handleLogout)
	http.HandleFunc("DELETE /_auth/session/{id}", auth.endSession)
}

func (auth AuthController) Handle(r *http.Request) congo.Controller {
	auth.Request = r
	return &auth
}

func (auth *AuthController) Current(roles ...string) *Identity {
	identity, _ := auth.CongoAuth.Authenticate(auth.Request, roles...)
	return identity
}

func (auth *AuthController) Usage() ([]*Usage, error) {
	identity := auth.Current(auth.PathValue("role"))
	return identity.Usages()
}

func (auth *AuthController) Identities() ([]*Identity, error) {
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

func (auth AuthController) handleSignup(w http.ResponseWriter, r *http.Request) {
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
			auth.Redirect(w, r, auth.CongoAuth.SetupRedirect)
			return
		}
		auth.Refresh(w, r)
	}
}

func (auth AuthController) handleSignin(w http.ResponseWriter, r *http.Request) {
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

func (auth AuthController) handleLogout(w http.ResponseWriter, r *http.Request) {
	role := r.PathValue("role")
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

	auth.Redirect(w, r, auth.CongoAuth.LogoutRedirect)
}

func (auth AuthController) endSession(w http.ResponseWriter, r *http.Request) {
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
