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

func (dir *Directory) Secure(fn congo.HandlerFunc, roles ...string) congo.HandlerFunc {
	return func(app *congo.Application, w http.ResponseWriter, r *http.Request) {
		if i, _ := dir.Authenticate(r); i != nil {
			if len(roles) == 0 {
				dir.TrackUsage(i, r.URL.String(), true)
				fn(app, w, r)
				return
			}
			for _, role := range roles {
				if i.Role == role {
					dir.TrackUsage(i, r.URL.String(), true)
					fn(app, w, r)
					return
				}
			}
			dir.TrackUsage(i, r.URL.String(), false)
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
}

func (dir *Directory) TrackUsage(i *Identity, url string, allowed bool) error {
	u := Usage{dir.DB.NewModel(uuid.NewString()), i.ID, url, allowed}
	return dir.DB.Query(`
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
