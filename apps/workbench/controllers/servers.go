package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/ccutch/congo/apps/workbench/models"
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

func (servers *ServersController) Servers() ([]*models.Server, error) {
	return models.AllServers(servers.DB)
}

func (servers ServersController) handleLaunch(w http.ResponseWriter, r *http.Request) {
	storage, err := strconv.Atoi(r.FormValue("storage"))
	if err != nil {
		servers.Render(w, r, "error-message", err)
		return
	}

	name, region, size := r.FormValue("name"), r.FormValue("region"), r.FormValue("size")
	server, err := models.NewServer(servers.DB, name, region, "")
	if err != nil {
		servers.Render(w, r, "error-message", err)
	}

	go func() {
		host, err := servers.hosting.NewServer(name, region, size, int64(storage))
		if err != nil {
			log.Println("Failed to start server")
			server.Error = err.Error()
		} else {
			server.IpAddress = host.IP
		}
		if err := server.Save(); err != nil {
			log.Println("Failed to save server", server, err)
		}

		if coding, ok := servers.Use("coding").(*CodingController); ok {
			if source, err := coding.Repository.Build("master", "."); err != nil {
				log.Println("Failed to build binary: ", err)
			} else if err = host.Deploy(source); err != nil {
				log.Println("Failed to deploy server: ", err)
			}
		}
	}()

	servers.Refresh(w, r)
}

func (servers ServersController) handleDomain(w http.ResponseWriter, r *http.Request) {
	server, err := models.GetServer(servers.DB, r.FormValue("server"))
	if err != nil {
		servers.Render(w, r, "error-message", err)
		return
	}

	host, err := servers.hosting.LoadServer(server.Name, server.Region)
	if err != nil {
		servers.Render(w, r, "error-message", err)
		return
	}

	domain := r.FormValue("domain")
	if host.RegisterDomain(domain); host.Error != nil {
		server.Error = host.Error.Error()
	} else {
		server.Domain = domain
	}

	if err := server.Save(); err != nil {
		log.Println("Failed to save server", server, err)
		return
	}

	servers.Refresh(w, r)

}

func (servers ServersController) handleRestart(w http.ResponseWriter, r *http.Request) {
	server, err := models.GetServer(servers.DB, r.PathValue("server"))
	if err != nil {
		servers.Render(w, r, "error-message", err)
		return
	}

	host, err := servers.hosting.LoadServer(server.Name, server.Region)
	if err != nil {
		servers.Render(w, r, "error-message", err)
		return
	}

	go func() {
		if coding, ok := servers.Use("coding").(*CodingController); ok {
			if source, err := coding.Repository.Build("master", "."); err != nil {
				log.Println("Failed to build binary: ", err)
			} else if err = host.Deploy(source); err != nil {
				log.Println("Failed to deploy server: ", err)
			}
		}
	}()

	servers.Refresh(w, r)
}
