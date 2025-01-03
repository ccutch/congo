package congo_auth

import (
	"cmp"
	"log"
	"net/http"
)

func (auth *Controller) Protect(h http.Handler, roles ...string) http.Handler {
	return auth.ProtectFunc(h.ServeHTTP, roles...)
}

func (app *Controller) ProtectFunc(fn http.HandlerFunc, roles ...string) http.HandlerFunc {
	if len(roles) == 0 {
		for role := range app.SigninViews {
			roles = append(roles, role)
		}
	}
	for _, role := range roles {
		if app.SigninViews[role] == "" {
			log.Fatal("Missing login view for role: ", role)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if app.CongoAuth.SetupView != "" && app.CongoAuth.count() == 0 {
			app.Render(w, r, app.CongoAuth.SetupView, nil)
			return
		}
		for _, role := range roles {
			if i, _ := app.CongoAuth.Authenticate(role, r); i != nil {
				// app.CongoAuth.NewUsage(i, r.URL.String(), true)
				fn(w, r)
				return
			}
		}
		view := app.CongoAuth.SigninViews[roles[0]]
		app.Render(w, r, cmp.Or(view, "congo-signin.html"), roles[0])
	}
}
