
package congo_book

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Chapter struct {
	congo.Model
	book   *CongoBook
	BookID string
	Number int
	Title  string
}

func (b *Book) NewChapter(orderNum int, title string) (*Chapter, error) {
	return b.NewChapterWithID(uuid.NewString(), orderNum, title)
}

func (b *Book) NewChapterWithID(id string, orderNum int, title string) (*Chapter, error) {
	s := Chapter{b.DB.NewModel(id), b.book, b.ID, orderNum, title}
	return &s, b.DB.Query(`

		INSERT INTO chapters (id, book_id, number, title)
		VALUES (?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, s.ID, s.BookID, s.Number, s.Title).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (b *Book) GetChapter(id string) (*Chapter, error) {
	s := Chapter{Model: b.DB.Model(), book: b.book}
	return &s, b.DB.Query(`

		SELECT id, book_id, number, title, created_at, updated_at
		FROM chapters
		WHERE id = ?

	`, id).Scan(&s.ID, &s.BookID, &s.Number, &s.Title, &s.CreatedAt, &s.UpdatedAt)
}

func (b *Book) Chapters() (chapters []*Chapter, err error) {
	return chapters, b.DB.Query(`

		SELECT id, book_id, number, title, created_at, updated_at
		FROM chapters
		WHERE book_id = $1
		ORDER BY number ASC

	`, b.ID).All(func(scan congo.Scanner) error {
		s := Chapter{Model: b.DB.Model(), book: b.book}
		chapters = append(chapters, &s)
		return scan(&s.ID, &s.BookID, &s.Number, &s.Title, &s.CreatedAt, &s.UpdatedAt)
	})
}

func (s *Chapter) Save() error {
	return s.DB.Query(`

		UPDATE chapters
		SET number = ?,
				title = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, s.Number, s.Title, s.ID).Scan(&s.UpdatedAt)
}

func (s *Chapter) Delete() error {
	return s.DB.Query(`

		DELETE FROM chapters
		WHERE id = ?

	`, s.ID).Exec()
}
