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

type CodingController struct {
	congo.BaseController
	host      *congo_host.CongoHost
	code      *congo_code.CongoCode
	Repo      *congo_code.Repository
	Workspace *congo_code.Workspace
}

func (coding *CodingController) Setup(app *congo.Application) {
	coding.BaseController.Setup(app)
	coding.host = congo_host.InitCongoHost(app.DB.Root)
	coding.code = congo_code.InitCongoCode(app.DB.Root)
	coding.Repo, _ = coding.code.NewRepo("code", congo_code.WithName("Code"))

	auth, ok := app.Use("auth").(*congo_auth.AuthController)
	if !ok {
		log.Fatal("Missing auth controller")
	}

	go func() {
		var err error
		coding.Workspace, err = coding.code.NewWorkspace(coding.host, "coder", 7000, coding.Repo)
		if err != nil {
			log.Println("Failed to setup workspace: ", err)
			return
		}

		if err = coding.Workspace.Start(); err != nil {
			log.Println("Failed to start workspace: ", err)
			return
		}
	}()

	http.HandleFunc("/_coding/download", auth.ProtectFunc(coding.handleDownload, "developer"))
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
