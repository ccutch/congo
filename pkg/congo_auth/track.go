package congo_auth

import "net/http"

func (auth *Controller) Track(h http.Handler, roles ...string) http.Handler {
	return auth.TrackFunc(h.ServeHTTP, roles...)
}

func (app *Controller) TrackFunc(fn http.HandlerFunc, roles ...string) http.HandlerFunc {
	if len(roles) == 0 {
		for role := range app.SigninViews {
			roles = append(roles, role)
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
				break
			}
		}
		fn(w, r)
	}
}
