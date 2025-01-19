package congo_auth

import (
	"cmp"
	"log"
	"net/http"
)

func (auth *AuthController) Serve(name string, roles ...string) http.Handler {
	if len(roles) == 0 {
		log.Print("Expecting roles if serving with auth controller.")
		log.Fatal("If you want to not restrict access then use app.Serve instead.")
	}
	return auth.Protect(auth.Application.Serve(name), roles...)
}

func (app *AuthController) Protect(h http.Handler, roles ...string) http.Handler {
	return app.ProtectFunc(h.ServeHTTP, roles...)
}

func (app *AuthController) ProtectFunc(fn http.HandlerFunc, roles ...string) http.HandlerFunc {
	if len(roles) == 0 {
		log.Print("Expecting roles if protecting an HTTP handler.")
		log.Fatal("If you want to not restrict access then use http.HandleFunc instead.")
	}
	for _, role := range roles {
		if v, ok := app.SigninViews[role]; !ok || v == "" {
			log.Fatal("Missing login view for role: ", role)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if app.CongoAuth.SetupView != "" && app.CongoAuth.Count() == 0 {
			app.Render(w, r, app.CongoAuth.SetupView, nil)
			return
		}
		if i, _ := app.CongoAuth.Authenticate(r, roles...); i != nil {
			fn(w, r)
			return
		}
		view := app.CongoAuth.SigninViews[roles[0]]
		app.Render(w, r, cmp.Or(view, "congo-signin.html"), roles[0])
	}
}
