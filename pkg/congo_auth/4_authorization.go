package congo_auth

import (
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

func (auth *CongoAuth) Protect(h http.Handler, roles ...string) http.HandlerFunc {
	return auth.ProtectFunc(h.ServeHTTP, roles...)
}

func (auth *CongoAuth) ProtectFunc(fn http.HandlerFunc, roles ...string) http.HandlerFunc {
	if len(roles) == 0 {
		roles = []string{auth.DefaultRole}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.SetupView != "" && auth.count() == 0 {
			auth.app.Render(w, r, auth.SetupView, nil)
			return
		}
		for _, role := range roles {
			if i, _ := auth.Authenticate(role, r); i != nil {
				auth.TrackUsage(i, r.URL.String(), true)
				fn(w, r)
				return
			}
		}
		if len(roles) == 1 {
			auth.app.Render(w, r, auth.LoginView, roles[0])
		} else {
			auth.app.Render(w, r, "congo-role-select.html", roles)
		}
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
	uarr := []*Usage{}
	return uarr, i.DB.Query(`
		SELECT id, identity_id, resource, allowed, created_at
		FROM usages
		WHERE identity_id = ?
		ORDER BY created_at DESC
	`, i.ID).All(func(scan congo.Scanner) error {
		u := Usage{Model: congo.Model{DB: i.DB}}
		uarr = append(uarr, &u)
		return scan(&u.ID, &u.IdentID, &u.Resource, &u.Allowed, &u.CreatedAt)
	})
}
