package congo_auth

import (
	"log"
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Usage struct {
	congo.Model
	IdentID  string
	Resource string
	Allowed  bool
}

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
				app.CongoAuth.TrackUsage(i, r.URL.String(), true)
				fn(w, r)
				return
			}
		}
		app.Render(w, r, app.CongoAuth.SigninViews[roles[0]], roles[0])
	}
}

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
				app.CongoAuth.TrackUsage(i, r.URL.String(), true)
				break
			}
		}
		fn(w, r)
	}
}

func (auth *CongoAuth) TrackUsage(i *Identity, url string, allowed bool) error {
	u := Usage{auth.DB.NewModel(uuid.NewString()), i.ID, url, allowed}
	return auth.DB.Query(`

		INSERT INTO usages (id, identity_id, resource, allowed)
		VALUES (?, ?, ?, ?)
		RETURNING created_at

	`, u.ID, u.IdentID, u.Resource, u.Allowed).Scan(&u.CreatedAt)
}

func (i *Identity) Usages() ([]*Usage, error) {
	usages := []*Usage{}
	return usages, i.DB.Query(`

		SELECT id, identity_id, resource, allowed, created_at
		FROM usages
		WHERE identity_id = ?
		ORDER BY created_at DESC

	`, i.ID).All(func(scan congo.Scanner) error {
		u := Usage{Model: congo.Model{DB: i.DB}}
		usages = append(usages, &u)
		return scan(&u.ID, &u.IdentID, &u.Resource, &u.Allowed, &u.CreatedAt)
	})
}
