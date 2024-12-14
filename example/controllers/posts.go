package controllers

import (
	"net/http"

	"github.com/ccutch/congo/example/models"
	"github.com/ccutch/congo/pkg/congo"
)

type PostController struct{ congo.BaseController }

func (ctrl *PostController) Setup(app *congo.Application) {
	ctrl.Application = app
	app.HandleFunc("POST /blog", ctrl.handleCreate)
	app.HandleFunc("PUT /blog/{post}", ctrl.handleUpdate)
}

func (ctrl PostController) Handle(r *http.Request) congo.Controller {
	ctrl.Request = r
	return &ctrl
}

func (app *PostController) CurrentPost() (*models.Post, error) {
	return models.GetPost(app.Application.DB, app.PathValue("post"))
}

func (app *PostController) SearchPosts() ([]*models.Post, error) {
	return models.SearchPosts(app.DB, app.PathValue("query"))
}

func (app PostController) handleCreate(w http.ResponseWriter, r *http.Request) {
	title, content := r.FormValue("title"), r.FormValue("content")
	post, err := models.NewPost(app.DB, title, content)
	if err != nil {
		app.Render(w, r, "error-message", err)
		return
	}
	app.Redirect(w, r, "/blog/"+post.ID)
}

func (app PostController) handleUpdate(w http.ResponseWriter, r *http.Request) {
	post, err := models.GetPost(app.DB, r.PathValue("post"))
	if err != nil {
		app.Render(w, r, "error-message", err)
		return
	}
	post.Title = r.FormValue("title")
	post.Content = r.FormValue("content")
	if err = post.Save(); err != nil {
		app.Render(w, r, "error-message", err)
		return
	}
	app.Redirect(w, r, "/blog/"+post.ID)
}
