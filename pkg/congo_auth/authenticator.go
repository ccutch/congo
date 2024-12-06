package congo_auth

import (
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
)

type Authenticator struct {
	DB *congo.Database
}

func NewAuthenticator(db *congo.Database) *Authenticator {
	return &Authenticator{db}
}

func (auth *Authenticator) Secure(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.Authenticate(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	}
}

func (auth *Authenticator) SecureFunc(fn congo.HandlerFunc) congo.HandlerFunc {
	return func(app *congo.Application, w http.ResponseWriter, r *http.Request) {
		if !auth.Authenticate(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		fn(app, w, r)
	}
}

func (auth *Authenticator) Authenticate(r *http.Request) bool {
	cookie, err := r.Cookie("academy-user")
	if err != nil {
		return false
	}
	if _, err = auth.ParseToken(cookie.Value); err != nil {
		return false
	}
	return true
}
