package congo_auth

import "net/http"

func (auth *AuthController) Check(h http.Handler, roles ...string) http.Handler {
	return auth.CheckFunc(h.ServeHTTP, roles...)
}

func (app *AuthController) CheckFunc(fn http.HandlerFunc, roles ...string) http.HandlerFunc {
	if len(roles) == 0 {
		for role := range app.SigninViews {
			roles = append(roles, role)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if app.CongoAuth.SetupView != "" && app.CongoAuth.Count() == 0 {
			app.Render(w, r, app.CongoAuth.SetupView, nil)
			return
		}
		fn(w, r)
	}
}
