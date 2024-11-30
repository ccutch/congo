package models

import (
	"congo.gitpost.app/internal/congo"
	"github.com/google/uuid"
)

type Post struct {
	congo.Model

	Title   string
	Content string
}

func NewPost(db *congo.Database, title, content string) (*Post, error) {
	post := Post{
		Model:   db.NewModel(uuid.NewString()),
		Title:   title,
		Content: content,
	}
	return &post, db.Query(`

		INSERT INTO posts (id, title, content)
		VALUES (?, ?, ?)
		RETURNING created_at, updated_at
	
	`, post.ID, post.Title, post.Content).Scan(&post.CreatedAt, &post.UpdatedAt)
}

func GetPost(db *congo.Database, id string) (*Post, error) {
	return &Post{
		Model:   congo.Model{ID: "test-post"},
		Title:   "Test Title",
		Content: "This is a test post...",
	}, nil
}

func SearchPosts(db *congo.Database, id string) ([]*Post, error) {
	return []*Post{
		{
			Model:   congo.Model{ID: "test-post"},
			Title:   "Test Title",
			Content: "This is a test post...",
		},
		{
			Model:   congo.Model{ID: "test-post"},
			Title:   "Test Title",
			Content: "This is a test post...",
		},
	}, nil
}
