package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_host"
)

type ServersController struct {
	congo.BaseController
	hosting *congo_host.Controller
}

func (servers *ServersController) Setup(app *congo.Application) {
	servers.BaseController.Setup(app)

	auth := app.Use("auth").(*congo_auth.Controller)
	servers.hosting = app.Use("hosting").(*congo_host.Controller)

	app.HandleFunc("POST /_servers/launch", auth.ProtectFunc(servers.handleLaunch))
	app.HandleFunc("POST /_servers/domain", auth.ProtectFunc(servers.handleDomain))
	app.HandleFunc("POST /_servers/restart/{server}", auth.ProtectFunc(servers.handleRestart))
}

func (servers ServersController) Handle(req *http.Request) congo.Controller {
	servers.Request = req
	return &servers
}

func (servers *ServersController) Servers() ([]*congo_host.Server, error) {
	return servers.hosting.ListServers()
}

func (servers ServersController) handleLaunch(w http.ResponseWriter, r *http.Request) {
	storage, err := strconv.Atoi(r.FormValue("storage"))
	if err != nil {
		servers.hosting.Render(w, r, "error-message", err)
		return
	}
	name, region, size := r.FormValue("name"), r.FormValue("region"), r.FormValue("size")
	server, err := servers.hosting.NewServer(name, region, size, int64(storage))
	if err != nil {
		servers.hosting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		if server.Setup(); server.Error != nil {
			server.Save()
			return
		}
		if coding, ok := servers.Use("coding").(*CodingController); ok {
			if source, err := coding.Repo.Build("master", "."); err != nil {
				log.Println("Failed to build binary: ", err)
			} else if err = server.Deploy(source); err != nil {
				log.Println("Failed to deploy server: ", err)
			}
		}
	}()

	servers.Refresh(w, r)
}

func (servers ServersController) handleDomain(w http.ResponseWriter, r *http.Request) {
	server, err := servers.hosting.LoadServer(r.FormValue("server"))
	if err != nil {
		servers.Render(w, r, "error-message", err)
		return
	}

	if domain := r.FormValue("domain"); domain != "" {
		if d, err := server.NewDomain(domain); err != nil {
			servers.Render(w, r, "error-message", err)
			return
		} else if err = d.Verify(); err == nil {
			d.Verified = true
			d.Save()
		}
	}

	if err := server.Save(); err != nil {
		servers.Render(w, r, "error-message", err)
		return
	}

	servers.Refresh(w, r)
}

func (servers ServersController) handleRestart(w http.ResponseWriter, r *http.Request) {
	server, err := servers.hosting.LoadServer(r.PathValue("server"))
	if err != nil {
		servers.Render(w, r, "error-message", err)
		return
	}

	host, err := servers.hosting.LoadServer(server.Name)
	if err != nil {
		servers.Render(w, r, "error-message", err)
		return
	}

	go func() {
		if coding, ok := servers.Use("coding").(*CodingController); ok {
			if source, err := coding.Repo.Build("master", "."); err != nil {
				log.Println("Failed to build binary: ", err)
			} else if err = host.Deploy(source); err != nil {
				log.Println("Failed to deploy server: ", err)
			}
		}
	}()

	servers.Refresh(w, r)
}
