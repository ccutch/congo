package controllers

import (
	"cmp"
	"log"
	"net/http"
	"path/filepath"

	"github.com/ccutch/congo/apps/workhouse/models"
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_code"
	"github.com/ccutch/congo/pkg/congo_host"
)

type ContentController struct {
	congo.BaseController

	Code *congo_code.CongoCode
	Host *congo_host.CongoHost
	Repo *congo_code.Repository
}

func (c *ContentController) Setup(app *congo.Application) {
	c.BaseController.Setup(app)

	c.Code = congo_code.InitCongoCode(app.DB.Root)
	c.Host = congo_host.InitCongoHost(app.DB.Root, nil)
	c.Repo, _ = c.Code.NewRepo("source", congo_code.WithName("Code"))

	auth := app.Use("auth").(*AuthController)
	app.Handle("/source/", c.Repo.Serve(auth.AuthController, "developer"))
	app.Handle("/coder/", auth.ProtectFunc(c.handleWorkspace, "developer"))
	app.Handle("/download", auth.ProtectFunc(c.downloadSource, "developer"))
	app.Handle("POST /_content/post", auth.ProtectFunc(c.publishPost, "developer"))
}

func (c ContentController) Handle(req *http.Request) congo.Controller {
	c.Request = req
	return &c
}

func (c *ContentController) Files() []*congo_code.Blob {
	branch := cmp.Or(c.URL.Query().Get("branch"), "master")
	blobs, _ := c.Repo.Blobs(branch, c.URL.Path)
	return blobs
}

func (c *ContentController) CurrentBranch() string {
	return cmp.Or(c.URL.Query().Get("branch"), "master")
}

func (c *ContentController) CurrentFile() (blob *congo_code.Blob, err error) {
	branch, path := c.CurrentBranch(), c.PathValue("path")
	if blob, err = c.Repo.Open(branch, path); err != nil {
		path = filepath.Join(path, "README.md")
		if blob, err = c.Repo.Open(branch, path); err != nil {
			return nil, nil
		}
	}
	return blob, err
}

func (c *ContentController) Posts() ([]*models.Post, error) {
	return models.AllPosts(c.DB)
}

func (c *ContentController) Hosts(ownerID string) []*congo_host.RemoteHost {
	// TODO use ownership table to filter servers
	hosts, err := c.Host.ListServers()
	log.Println(hosts, err)
	return nil
}

func (c ContentController) downloadSource(w http.ResponseWriter, r *http.Request) {
	branch := cmp.Or(r.URL.Query().Get("branch"), "master")
	path := cmp.Or(r.URL.Query().Get("path"), ".")
	binary, err := c.Repo.Build(branch, path)
	if err != nil {
		log.Println("Failed to build binary: ", err)
	}
	w.Header().Set("Content-Disposition", "attachment; filename=congo")
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, binary)
}

func (c ContentController) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	i, _ := c.Use("auth").(*AuthController).Authenticate(r, "developer")
	if workspace, err := c.Code.GetWorkspace("workspace-" + i.ID); err == nil {
		workspace.Proxy(r.URL.Path).ServeHTTP(w, r)
		return
	}
	c.Render(w, r, "not-found.html", nil)
}

func (c ContentController) publishPost(w http.ResponseWriter, r *http.Request) {
	_, err := models.NewPost(c.DB, r.FormValue("title"), r.FormValue("content"))
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	c.Refresh(w, r)
}
