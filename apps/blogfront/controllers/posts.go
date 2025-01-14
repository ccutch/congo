package controllers

import (
	"net/http"

	"github.com/ccutch/congo/apps/blogfront/models"
	"github.com/ccutch/congo/pkg/congo"
)

type PostController struct{ congo.BaseController }

func (posts *PostController) Setup(app *congo.Application) {
	posts.Application = app
	app.HandleFunc("POST /blog", posts.handleCreate)
	app.HandleFunc("PUT /blog/{post}", posts.handleUpdate)
}

func (posts PostController) Handle(r *http.Request) congo.Controller {
	posts.Request = r
	return &posts
}

func (posts *PostController) CurrentPost() (*models.Post, error) {
	return models.GetPost(posts.Application.DB, posts.PathValue("post"))
}

func (posts *PostController) SearchPosts() ([]*models.Post, error) {
	return models.SearchPosts(posts.DB, posts.PathValue("query"))
}

func (posts PostController) handleCreate(w http.ResponseWriter, r *http.Request) {
	title, content := r.FormValue("title"), r.FormValue("content")
	post, err := models.NewPost(posts.DB, title, content)
	if err != nil {
		posts.Render(w, r, "error-message", err)
		return
	}
	posts.Redirect(w, r, "/blog/"+post.ID)
}

func (posts PostController) handleUpdate(w http.ResponseWriter, r *http.Request) {
	post, err := models.GetPost(posts.DB, r.PathValue("post"))
	if err != nil {
		posts.Render(w, r, "error-message", err)
		return
	}
	post.Title = r.FormValue("title")
	post.Content = r.FormValue("content")
	if err = post.Save(); err != nil {
		posts.Render(w, r, "error-message", err)
		return
	}
	posts.Redirect(w, r, "/blog/"+post.ID)
}
