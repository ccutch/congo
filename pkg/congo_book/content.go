package congo_book

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Content struct {
	congo.Model
	book      *CongoBook
	BookID    string
	ChapterID string
	Position  int
	MediaType string
	Content   []byte
}

func (b *Book) NewContent(position int, media string, content []byte) (*Content, error) {
	c := Content{b.DB.NewModel(uuid.NewString()), b.book, b.ID, "", position, media, content}
	return &c, b.DB.Query(`

		INSERT INTO contents (id, book_id, chapter_id, position, media, content)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, c.ID, c.BookID, c.ChapterID, c.Position, c.MediaType, c.Content).Scan(&c.CreatedAt, &c.UpdatedAt)
}

func (b *Book) Contents() ([]*Content, error) {
	var contents []*Content
	return contents, b.DB.Query(`

		SELECT id, book_id, chapter_id, position, media, content, created_at, updated_at
		FROM contents
		WHERE book_id = ? AND chapter_id = ''
		ORDER BY created_at DESC

	`, b.ID).All(func(scan congo.Scanner) error {
		c := Content{Model: b.DB.Model(), book: b.book}
		contents = append(contents, &c)
		return scan(&c.ID, &c.BookID, &c.ChapterID, &c.Position, &c.Content, &c.CreatedAt, &c.UpdatedAt)
	})
}

func (ch *Chapter) NewContent(position int, media string, content []byte) (*Content, error) {
	c := Content{ch.DB.NewModel(uuid.NewString()), ch.book, ch.BookID, ch.ID, position, media, content}
	return &c, c.DB.Query(`

		INSERT INTO contents (id, book_id, chapter_id, position, media, content)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, c.ID, c.BookID, c.ChapterID, c.Position, c.MediaType, c.Content).Scan(&c.CreatedAt, &c.UpdatedAt)
}

func (ch *Chapter) Contents() ([]*Content, error) {
	var contents []*Content
	return contents, ch.DB.Query(`

		SELECT id, book_id, chapter_id, position, media, content, created_at, updated_at
		FROM contents
		WHERE book_id = ? AND chapter_id = ?
		ORDER BY created_at DESC

	`, ch.BookID, ch.ID).All(func(scan congo.Scanner) error {
		c := Content{Model: ch.DB.Model(), book: ch.book}
		contents = append(contents, &c)
		return scan(&c.ID, &c.BookID, &c.ChapterID, &c.Position, &c.MediaType, &c.Content, &c.CreatedAt, &c.UpdatedAt)
	})
}

func (c *Content) Save() error {
	return c.DB.Query(`

		UPDATE contents
		SET content = ?,
				media = ?,
		    position = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, c.Content, c.MediaType, c.Position, c.ID).Scan(&c.UpdatedAt)
}

func (c *Content) Delete() error {
	return c.DB.Query(`

		DELETE FROM contents
		WHERE id = ?

	`, c.ID).Exec()
}
