package controllers

import (
	"cmp"
	"log"
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_code"
	"github.com/ccutch/congo/pkg/congo_host"
)

func Coding(host *congo_host.CongoHost, code *congo_code.CongoCode) (string, *CodingController) {
	return "coding", &CodingController{host: host, code: code}
}

type CodingController struct {
	congo.BaseController
	host      *congo_host.CongoHost
	code      *congo_code.CongoCode
	Repo      *congo_code.Repository
	Workspace *congo_code.Workspace
}

func (coding *CodingController) Setup(app *congo.Application) {
	coding.BaseController.Setup(app)
	coding.Repo, _ = coding.code.NewRepo("code", congo_code.WithName("Code"))
	coding.Workspace, _ = coding.code.NewWorkspace(coding.host, "coder", 7000, coding.Repo)

	go func() {
		if err := coding.Workspace.Start(); err != nil {
			log.Println("Failed to start workspace: ", err)
			return
		}
	}()

	auth := app.Use("auth").(*congo_auth.AuthController)
	http.Handle("/raw/{path...}", auth.ProtectFunc(coding.handleRaw, "developer"))
	http.Handle("/_coding/download", auth.ProtectFunc(coding.handleDownload, "developer"))
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

func (coding *CodingController) handleDownload(w http.ResponseWriter, r *http.Request) {
	path, err := coding.Repo.Build("master", ".")
	if err != nil {
		log.Println("Failed to build binary: ", err)
	}
	w.Header().Set("Content-Disposition", "attachment; filename=congo")
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, path)
}

func (coding *CodingController) handleRaw(w http.ResponseWriter, r *http.Request) {
	blob, err := coding.Repo.Open("master", r.PathValue("path"))
	if err != nil {
		coding.Render(w, r, "error-message", err)
		return
	}
	content, err := blob.Content()
	if err != nil {
		coding.Render(w, r, "error-message", err)
		return
	}
	w.Header().Set("Content-Type", blob.FileType())
	w.Write([]byte(content))
}
