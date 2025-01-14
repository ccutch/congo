package congo_book

import (
	"time"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Book struct {
	congo.Model
	book      *CongoBook
	LibraryID string
	Public    bool
	Title     string
	Author    string
	PublishAt time.Time
}

func (l *Library) NewBook(title, author string, public bool) (*Book, error) {
	return l.NewBookWithID(uuid.NewString(), title, author, public)
}

func (l *Library) NewBookWithID(id, title, author string, public bool) (*Book, error) {
	b := Book{l.DB.NewModel(id), l.book, l.ID, public, title, author, time.Now()}
	return &b, l.DB.Query(`

		INSERT INTO books (id, library_id, public, title, author, publish_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, b.ID, b.LibraryID, b.Public, b.Title, b.Author, b.PublishAt).Scan(&b.CreatedAt, &b.UpdatedAt)
}

func (l *Library) GetBook(id string) (*Book, error) {
	b := Book{Model: l.DB.Model(), book: l.book}
	return &b, l.DB.Query(`

		SELECT id, library_id, public, title, author, publish_at, created_at, updated_at
		FROM books
		WHERE id = ?

	`, id).Scan(&b.ID, &b.LibraryID, &b.Public, &b.Title, &b.Author, &b.PublishAt, &b.CreatedAt, &b.UpdatedAt)
}

func (l *Library) SearchBooks(query string) (books []*Book, err error) {
	return books, l.DB.Query(`

		SELECT id, library_id, public, title, author, publish_at, created_at, updated_at
		FROM books
		WHERE library_id = $1 AND (id LIKE $2 OR title LIKE $2 OR author LIKE $2)

	`, l.ID, "%"+query+"%").All(func(scan congo.Scanner) error {
		b := Book{Model: l.DB.Model(), book: l.book}
		books = append(books, &b)
		return scan(&b.ID, &b.LibraryID, &b.Public, &b.Title, &b.Author, &b.PublishAt, &b.CreatedAt, &b.UpdatedAt)
	})
}

func (b *Book) Library() (*Library, error) {
	return b.book.GetLibrary(b.LibraryID)
}

func (b *Book) Save() error {
	return b.DB.Query(`

		UPDATE books
		SET title = ?,
				author = ?,
				public = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, b.Title, b.Author, b.Public, b.ID).Scan(&b.UpdatedAt)
}

func (b *Book) Delete() error {
	return b.DB.Query(`

		DELETE FROM books
		WHERE id = ?

	`, b.ID).Exec()
}
