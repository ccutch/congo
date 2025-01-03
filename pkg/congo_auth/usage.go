package congo_auth

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Usage struct {
	congo.Model
	IdentID  string
	Resource string
	Allowed  bool
}

func (auth *CongoAuth) NewUsage(i *Identity, url string, allowed bool) error {
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
