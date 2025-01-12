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

type CodingController struct {
	congo.BaseController
	Host *congo_host.CongoHost
	Code *congo_code.CongoCode
	Repo *congo_code.Repository
}

func (coding *CodingController) Setup(app *congo.Application) {
	var err error

	coding.BaseController.Setup(app)
	coding.Repo, err = coding.Code.NewRepo("code", congo_code.WithName("Code"))
	if err != nil {
		log.Fatal("Failed to create repo: ", err)
	}

	auth := app.Use("auth").(*congo_auth.Controller)
	app.Handle("/code/", coding.Repo.Serve(auth, "developer"))
	app.Handle("/@{user}/", auth.ProtectFunc(coding.handleWorkspace))
	app.Handle("/_coding/download", auth.ProtectFunc(coding.handleDownload))
}

func (coding CodingController) Handle(req *http.Request) congo.Controller {
	coding.Request = req
	return &coding
}

func (coding *CodingController) Files() []*congo_code.Blob {
	branch := cmp.Or(coding.URL.Query().Get("branch"), "master")
	blobs, _ := coding.Repo.Blobs(branch, coding.URL.Path)
	return blobs
}

func (coding *CodingController) CurrentFile() *congo_code.Blob {
	branch := cmp.Or(coding.URL.Query().Get("branch"), "master")
	blob, err := coding.Repo.Open(branch, coding.URL.Path[1:])
	if err != nil {
		return nil
	}
	return blob
}

func (coding *CodingController) HandleNewSignup(auth *congo_auth.Controller, i *congo_auth.Identity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, _ := coding.Code.AllWorkspaces()
		workspace, err := coding.Code.RunWorkspace(coding.Host, "workspace-"+i.Name, 7000+len(ws), coding.Repo)
		if err != nil {
			auth.Render(w, r, "error-message", err)
			return
		}

		go workspace.Start()
		auth.Redirect(w, r, "/@"+i.Name)
	}
}

func (coding *CodingController) handleDownload(w http.ResponseWriter, r *http.Request) {
	path, err := coding.Repo.Build("master", ".")
	if err != nil {
		log.Println("Failed to build binary: ", err)
	}
	w.Header().Set("Content-Disposition", "attachment; filename=congo")
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, path)
}

func (coding *CodingController) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := fmt.Sprintf("workspace-%s", r.PathValue("user"))
	workspace, err := coding.Code.GetWorkspace(workspaceID)
	if err != nil {
		n, _ := coding.Code.AllWorkspaces()
		workspace, err = coding.Code.RunWorkspace(coding.Host, workspaceID, 7000+len(n), coding.Repo)
		if err != nil {
			coding.Render(w, r, "error-message", err)
			return
		}

		go workspace.Start()
	}

	workspace.Proxy(r.URL.Path).ServeHTTP(w, r)
}
