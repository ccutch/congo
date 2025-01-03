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

type HostingController struct {
	congo.BaseController
	host *congo_host.Controller
}

func (hosting *HostingController) Setup(app *congo.Application) {
	hosting.BaseController.Setup(app)

	auth := app.Use("auth").(*congo_auth.Controller)
	hosting.host = app.Use("host").(*congo_host.Controller)

	app.HandleFunc("POST /_hosting/launch", auth.ProtectFunc(hosting.handleLaunch))
	app.HandleFunc("POST /_hosting/domain", auth.ProtectFunc(hosting.handleDomain))
	app.HandleFunc("POST /_hosting/restart/{server}", auth.ProtectFunc(hosting.handleRestart))
}

func (hosting HostingController) Handle(req *http.Request) congo.Controller {
	hosting.Request = req
	return &hosting
}

func (hosting *HostingController) Servers() ([]*models.Server, error) {
	return models.AllServers(hosting.DB)
}

func (hosting HostingController) handleLaunch(w http.ResponseWriter, r *http.Request) {
	storage, err := strconv.Atoi(r.FormValue("storage"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	name, region, size := r.FormValue("name"), r.FormValue("region"), r.FormValue("size")
	server, err := models.NewServer(hosting.DB, name, region, "")
	if err != nil {
		hosting.Render(w, r, "error-message", err)
	}

	go func() {
		host, err := hosting.host.NewServer(name, region, size, int64(storage))
		if err != nil {
			log.Println("Failed to start server")
			server.Error = err.Error()
		} else {
			server.IpAddress = host.IP
		}
		if err := server.Save(); err != nil {
			log.Println("Failed to save server", server, err)
		}

		if coding, ok := hosting.Use("coding").(*CodingController); ok {
			if source, err := coding.Repo.Build("master", "."); err != nil {
				log.Println("Failed to build binary: ", err)
			} else if err = host.Deploy(source); err != nil {
				log.Println("Failed to deploy server: ", err)
			}
		}
	}()

	hosting.Refresh(w, r)
}

func (hosting HostingController) handleDomain(w http.ResponseWriter, r *http.Request) {
	server, err := models.GetServer(hosting.DB, r.FormValue("server"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	host, err := hosting.host.LoadServer(server.Name, server.Region)
	if err != nil {
		hosting.Render(w, r, "error-message", err)
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

	hosting.Refresh(w, r)

}

func (hosting HostingController) handleRestart(w http.ResponseWriter, r *http.Request) {
	server, err := models.GetServer(hosting.DB, r.PathValue("server"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	host, err := hosting.host.LoadServer(server.Name, server.Region)
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		if coding, ok := hosting.Use("coding").(*CodingController); ok {
			if source, err := coding.Repo.Build("master", "."); err != nil {
				log.Println("Failed to build binary: ", err)
			} else if err = host.Deploy(source); err != nil {
				log.Println("Failed to deploy server: ", err)
			}
		}
	}()

	hosting.Refresh(w, r)
}
