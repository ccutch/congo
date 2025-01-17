package controllers

import (
	"cmp"
	"fmt"
	"log"
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_code"
	"github.com/ccutch/congo/pkg/congo_host"
)

type ContentController struct {
	congo.BaseController
	Host *congo_host.CongoHost
	Code *congo_code.CongoCode
	Repo *congo_code.Repository
}

func (content *ContentController) Setup(app *congo.Application) {
	auth := app.Use("auth").(*congo_auth.Controller)

	content.Host = congo_host.InitCongoHost(app.DB.Root, nil)
	content.Code = congo_code.InitCongoCode(app.DB.Root)
	content.Repo, _ = content.Code.NewRepo("code", congo_code.WithName("Code"))

	content.BaseController.Setup(app)
	app.Handle("/code/", content.Repo.Serve(auth, "developer"))
	app.Handle("/coder/", auth.ProtectFunc(content.handleWorkspace, "developer"))
	app.Handle("/_download", auth.ProtectFunc(content.handleDownload, "developer"))
}

func (content ContentController) Handle(req *http.Request) congo.Controller {
	content.Request = req
	return &content
}

func (content *ContentController) Files() []*congo_code.Blob {
	branch := cmp.Or(content.URL.Query().Get("branch"), "master")
	blobs, _ := content.Repo.Blobs(branch, content.URL.Path)
	return blobs
}

func (content *ContentController) CurrentFile() *congo_code.Blob {
	branch := cmp.Or(content.URL.Query().Get("branch"), "master")
	blob, err := content.Repo.Open(branch, content.URL.Path[1:])
	if err != nil {
		return nil
	}
	return blob
}

func (content ContentController) handleDownload(w http.ResponseWriter, r *http.Request) {
	path, err := content.Repo.Build("master", ".")
	if err != nil {
		log.Println("Failed to build binary: ", err)
	}
	w.Header().Set("Content-Disposition", "attachment; filename=congo")
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, path)
}

func (content ContentController) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	auth := content.Use("auth").(*congo_auth.Controller)

	i, _ := auth.Authenticate("developer", r)
	workspaceID := fmt.Sprintf("workspace-%s", i.ID)
	workspace, err := content.Code.GetWorkspace(workspaceID)
	if err != nil {
		content.Render(w, r, "not-found.html", nil)
		return
	}

	workspace.Proxy(r.URL.Path).ServeHTTP(w, r)
}
