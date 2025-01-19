package models

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Post struct {
	congo.Model
	Title   string
	Content string
}

func NewPost(db *congo.Database, title, content string) (*Post, error) {
	p := Post{db.NewModel(uuid.NewString()), title, content}
	return &p, db.Query(`

		INSERT INTO posts (id, title, content)
		VALUES (?, ?, ?)
		RETURNING created_at, updated_at

	`, p.ID, p.Title, p.Content).Scan(&p.CreatedAt, &p.UpdatedAt)
}

func AllPosts(db *congo.Database) (posts []*Post, err error) {
	return posts, db.Query(`

		SELECT id, title, content, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		
	`).All(func(scan congo.Scanner) error {
		p := Post{Model: db.Model()}
		posts = append(posts, &p)
		return scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt)
	})
}
