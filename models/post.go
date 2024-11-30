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
	post := Post{Model: congo.Model{Database: db}}
	return &post, db.Query(`
	
		SELECT id, title, content, created_at, updated_at
		FROM posts
		WHERE id = ?
	
	`, id).Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
}

func SearchPosts(db *congo.Database, query string) ([]*Post, error) {
	var posts []*Post
	return posts, db.Query(`
	
		SELECT id, title, content, created_at, updated_at
		FROM posts
		WHERE title LIKE ?
	
	`, "%"+query+"%").All(func(scan congo.Scanner) error {
		post := Post{Model: congo.Model{Database: db}}
		posts = append(posts, &post)
		return scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
	})
}

func (post *Post) Save() error {
	return post.Query(`
	
		UPDATE posts
		SET title = ?, content = ?
		WHERE id = ?
		RETURNING created_at, updated_at

	`, post.Title, post.Content, post.ID).Scan(&post.CreatedAt, &post.UpdatedAt)
}

func (post *Post) Delete() error {
	return post.Query(`
	
		DELETE FROM posts
		WHERE id = ?

	`, post.ID).Exec()
}