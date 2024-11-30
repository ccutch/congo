package controllers

import (
	"net/http"

	"congo.gitpost.app/internal/congo"
	"congo.gitpost.app/models"
)

type PostController struct{ congo.BaseController }

func (ctrl *PostController) Mount(server *congo.Server) error {
	ctrl.Server = server
	server.WithEndpoint("POST /blog", false, ctrl.CreatePost)
	server.WithEndpoint("PUT /blog/{post}", false, ctrl.UpdatePost)
	return nil
}

func (ctrl PostController) WithRequest(r *http.Request) congo.Controller {
	ctrl.Request = r
	return &ctrl
}

func (app *PostController) Current() (*models.Post, error) {
	return models.GetPost(app.Database, app.PathValue("post"))
}

func (app *PostController) Search() ([]*models.Post, error) {
	return models.SearchPosts(app.Database, app.PathValue("query"))
}

func (app PostController) CreatePost(w http.ResponseWriter, r *http.Request) {
	title, content := r.FormValue("title"), r.FormValue("content")
	post, err := models.NewPost(app.Database, title, content)
	if err != nil {
		app.Render(w, "error-message", err)
		return
	}
	app.Redirect(w, r, "/blog/"+post.ID)
}

func (app PostController) UpdatePost(w http.ResponseWriter, r *http.Request) {
	post, err := models.GetPost(app.Database, r.PathValue("post"))
	if err != nil {
		app.Render(w, "error-message", err)
		return
	}
	post.Title = r.FormValue("title")
	post.Content = r.FormValue("content")
	if err = post.Save(); err != nil {
		app.Render(w, "error-message", err)
		return
	}
	app.Redirect(w, r, "/blog/"+post.ID)
}
