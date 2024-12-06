package congo_auth

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type User struct {
	congo.Model
	Name     string
	Email    string
	Password string
}

func (auth *Authenticator) GetUser(id string) (*User, error) {
	u := User{Model: auth.DB.NewModel(uuid.NewString())}
	return &u, auth.DB.Query(`

		SELECT id, name, email, password, created_at, updated_at
		FROM users
		WHERE id = $1 OR name = $1 OR email = $1
	
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt)
}
