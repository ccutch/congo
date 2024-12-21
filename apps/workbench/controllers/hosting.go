package controllers

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_host"
)

type HostingController struct {
	congo.BaseController
	host *congo_host.CongoHost
}

func (hosting *HostingController) Setup(app *congo.Application) {
	hosting.BaseController.Setup(app)
	hosting.host = congo_host.InitCongoHost(filepath.Join(app.DB.Root, "hosts"))
	auth := congo_auth.InitCongoAuth(app, congo_auth.WithDefaultRole("developer"))
	app.HandleFunc("POST /_hosting/launch", auth.ProtectFunc(hosting.handleLaunch))
}

func (hosting HostingController) Handle(req *http.Request) congo.Controller {
	hosting.Request = req
	return &hosting
}

func (hosting *HostingController) Servers() ([]*congo_host.Server, error) {
	return hosting.host.ListServers()
}

func (hosting HostingController) handleLaunch(w http.ResponseWriter, r *http.Request) {
	storage, err := strconv.Atoi(r.FormValue("storage"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	name, region, size := r.FormValue("name"), r.FormValue("region"), r.FormValue("size")
	server, err := hosting.host.NewServer(name, region, size, int64(storage))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	go server.Start()
	hosting.Refresh(w, r)
}
