package controllers

import (
	"cmp"
	"log"
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_code"
)

type CodingController struct {
	congo.BaseController
	*congo_code.CongoCode
	Repo *congo_code.Repository
	Work *congo_code.Workspace
}

func (code *CodingController) Setup(app *congo.Application) {
	code.BaseController.Setup(app)
	code.CongoCode = congo_code.InitCongoCode(app.DB.Root)
	app.HandleFunc("/_coding/download", code.handleDownload)
}

func (code CodingController) Handle(req *http.Request) congo.Controller {
	code.Request = req
	return &code
}

func (code *CodingController) Files() []*congo_code.Blob {
	branch := cmp.Or(code.URL.Query().Get("branch"), "master")
	blobs, _ := code.Repo.Blobs(branch, code.URL.Path)
	return blobs
}

func (code *CodingController) CurrentFile() *congo_code.Blob {
	branch := cmp.Or(code.URL.Query().Get("branch"), "master")
	blob, err := code.Repo.Open(branch, code.URL.Path[1:])
	if err != nil {
		return nil
	}
	return blob
}

func (code *CodingController) handleDownload(w http.ResponseWriter, r *http.Request) {
	path, err := code.Repo.Build("master", ".")
	if err != nil {
		log.Println("Failed to build binary: ", err)
	}

	w.Header().Set("Content-Disposition", "attachment; filename=congo")
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, path)
}
