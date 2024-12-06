package congo_auth

import (
	"net/http"
	"time"
)

func (auth *Authenticator) StartSession(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "academy-user",
		Path:     "/",
		Value:    value,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
}

func (auth *Authenticator) StopSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "academy-user",
		Path:     "/",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // Set the expiration to the past to delete the cookie
		HttpOnly: true,
	})
}
