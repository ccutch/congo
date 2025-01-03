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
	post := Post{Model: congo.Model{DB: db}}
	return &post, db.Query(`
	
		SELECT id, title, content, created_at, updated_at
		FROM posts
		WHERE id = ?
	
	`, id).Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
}

func SearchPosts(db *congo.Database, query string) (posts []*Post, err error) {
	return posts, db.Query(`
	
		SELECT id, title, content, created_at, updated_at
		FROM posts
		WHERE title LIKE ?
	
	`, "%"+query+"%").All(func(scan congo.Scanner) (err error) {
		post := Post{Model: congo.Model{DB: db}}
		posts = append(posts, &post)
		return scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
	})
}

func (post *Post) Save() error {
	return post.DB.Query(`
	
		UPDATE posts
		SET title = ?, content = ?
		WHERE id = ?
		RETURNING created_at, updated_at

	`, post.Title, post.Content, post.ID).Scan(&post.CreatedAt, &post.UpdatedAt)
}

func (post *Post) Delete() error {
	return post.DB.Query(`
	
		DELETE FROM posts
		WHERE id = ?

	`, post.ID).Exec()
}
