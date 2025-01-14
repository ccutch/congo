package congo_book

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/google/uuid"
)

type Library struct {
	congo.Model
	book        *CongoBook
	OwnerID     string
	Public      bool
	Name        string
	Description string
}

func (book *CongoBook) NewLibrary(ownerID, name, description string, public bool) (*Library, error) {
	return book.NewLibraryWithID(uuid.NewString(), ownerID, name, description, public)
}

func (book *CongoBook) NewLibraryWithID(id, ownerID, name, description string, public bool) (*Library, error) {
	l := Library{book.db.NewModel(id), book, ownerID, public, name, description}
	return &l, book.db.Query(`

		INSERT INTO libraries (id, owner_id, public, name, description)
		VALUES (?, ?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, l.ID, l.OwnerID, l.Public, l.Name, l.Description).Scan(&l.CreatedAt, &l.UpdatedAt)
}

func (book *CongoBook) GetLibrary(id string) (*Library, error) {
	l := Library{Model: book.db.Model(), book: book}
	return &l, book.db.Query(`

		SELECT id, owner_id, public, name, description, created_at, updated_at
		FROM libraries
		WHERE id = ?

	`, id).Scan(&l.ID, &l.OwnerID, &l.Public, &l.Name, &l.Description, &l.CreatedAt, &l.UpdatedAt)
}

func (book *CongoBook) SearchLibraries(query string) (libraries []*Library, err error) {
	return libraries, book.db.Query(`

		SELECT id, owner_id, public, name, description, created_at, updated_at
		FROM libraries
		WHERE id LIKE $1 OR name LIKE $1 OR description LIKE $1

	`, "%"+query+"%").All(func(scan congo.Scanner) error {
		l := Library{Model: book.db.Model(), book: book}
		libraries = append(libraries, &l)
		return scan(&l.ID, &l.OwnerID, &l.Public, &l.Name, &l.Description, &l.CreatedAt, &l.UpdatedAt)
	})
}

func (l *Library) Owner() (*congo_auth.Identity, error) {
	return l.book.auth.Lookup(l.OwnerID)
}

func (l *Library) Save() error {
	return l.DB.Query(`

		UPDATE libraries
		SET name = ?,
				description = ?,
				public = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, l.Name, l.Description, l.Public, l.ID).Scan(&l.UpdatedAt)
}
