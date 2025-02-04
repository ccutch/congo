package controllers

import (
	"fmt"
	"net/http"

	"github.com/ccutch/congo/apps"
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_sell"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type HostingController struct {
	congo.BaseController
	Host *congo_host.CongoHost
	Sell *congo_sell.CongoSell
}

func (hosting *HostingController) Setup(app *congo.Application) {
	hosting.BaseController.Setup(app)
	app.HandleFunc("POST /launch", hosting.launchServer)
}

func (hosting HostingController) Handle(req *http.Request) congo.Controller {
	hosting.Request = req
	return &hosting
}

func (hosting HostingController) launchServer(w http.ResponseWriter, r *http.Request) {
	var (
		app    = r.FormValue("server-type")
		name   = fmt.Sprintf("demo-%s-%s", app, uuid.NewString())
		size   = "s-1vcpu-2gb"
		region = "sfo2"
	)

	server, err := hosting.Host.NewServer(name, size, region)
	if err != nil {
		hosting.Render(w, r, "error-message", errors.Wrap(err, "failed to init new server"))
		return
	}

	if err = server.Launch(region, size, 5); err != nil {
		hosting.Render(w, r, "error-message", errors.Wrap(err, "failed to launch new server"))
		return
	}

	if err = server.Prepare(); err != nil {
		hosting.Render(w, r, "error-message", errors.New("failed to prepare server"))
		return
	}

	dest, err := apps.Build(app)
	if err != nil {
		hosting.Render(w, r, "error-message", errors.Wrap(err, "failed to build source code"))
		return
	}

	if err = server.Deploy(dest); err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	hosting.Render(w, r, "success-link", struct {
		Server *congo_host.RemoteHost
		Name   string
	}{server, app})
}
