package controllers

import (
	"net/http"

	"congo.gitpost.app/internal/congo"
	"congo.gitpost.app/models"
)

type PostController struct {
	congo.BaseController
	*congo.Database
}

func (ctrl *PostController) Mount(server *congo.Server) error {
	ctrl.Database = server.Database
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
	title, content := app.FormValue("title"), app.FormValue("content")
	post, err := models.NewPost(app.Database, title, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	app.Redirect(w, r, "/blog/"+post.ID)
}
