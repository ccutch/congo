package congo_auth

import (
	"errors"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Identity struct {
	congo.Model
	Role     string
	Email    string
	Username string
	PassHash []byte
}

func hash(pass string) (hash []byte, err error) {
	if len(pass) == 0 {
		return nil, errors.New("expected password; recieved empty")
	}
	return bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
}

func (dir *Directory) Create(role, email, name, password string) (i *Identity, err error) {
	i = &Identity{Model: dir.DB.NewModel(uuid.NewString()), Role: role, Email: email, Username: name}
	if i.PassHash, err = hash(password); err != nil {
		return nil, err
	}
	return i, i.DB.Query(`
		INSERT INTO identities (id, role, email, username, passhash)
		VALUES (?, ?, ?, ?, ?)
		RETURNING created_at, updated_at
	`, i.ID, i.Role, i.Email, i.Username, i.PassHash).Scan(&i.CreatedAt, &i.UpdatedAt)
}

func (dir *Directory) Lookup(ident string) (*Identity, error) {
	i := &Identity{Model: congo.Model{DB: dir.DB}}
	return i, i.DB.Query(`
		SELECT id, role, email, username, passhash, created_at, updated_at
		FROM identities
		WHERE id = $1 OR email = $1 OR username = $1
	`, ident).Scan(&i.ID, &i.Role, &i.Email, &i.Username, &i.PassHash, &i.CreatedAt, &i.UpdatedAt)
}

func (dir *Directory) Search(query string) (imap map[string][]*Identity, err error) {
	imap = map[string][]*Identity{}
	return imap, dir.DB.Query(`
		SELECT id, role, email, username, passhash, created_at, updated_at
		FROM identities
		WHERE id LIKE $1 OR email LIKE $1 OR username LIKE $1
	`, "%"+query+"%").All(func(scan congo.Scanner) error {
		i := &Identity{Model: congo.Model{DB: dir.DB}}
		err = scan(&i.ID, &i.Role, &i.Email, &i.Username, &i.PassHash, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return err
		}
		imap[i.Role] = append(imap[i.Role], i)
		return nil
	})
}

func (dir *Directory) SearchByRole(role, query string) (iarr []*Identity, err error) {
	iarr = []*Identity{}
	return iarr, dir.DB.Query(`
		SELECT id, role, email, username, passhash, created_at, updated_at
		FROM identities
		WHERE role = $1 AND (id LIKE $2 OR email LIKE $2 OR username LIKE $2)
	`, role, "%"+query+"%").All(func(scan congo.Scanner) error {
		i := Identity{Model: congo.Model{DB: dir.DB}}
		iarr = append(iarr, &i)
		return scan(&i.ID, &i.Role, &i.Email, &i.Username, &i.PassHash, &i.CreatedAt, &i.UpdatedAt)
	})
}

func (i *Identity) Verify(password string) bool {
	return bcrypt.CompareHashAndPassword(i.PassHash, []byte(password)) == nil
}

func (i *Identity) Save() error {
	return i.DB.Query(`
		UPDATE identities
		SET role = ?, email = ?, username = ?, passhash = ?, updated_at = CURRENT_TIMESTAMP
		RETURNING updated_at
	`, i.Role, i.Email, i.Username, i.PassHash).Scan(&i.UpdatedAt)
}

func (i *Identity) Delete() error {
	return i.DB.Query(`
		DELETE FROM identities
		WHERE id = ?
	`, i.ID).Exec()
}
