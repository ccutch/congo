package controllers

import (
	"log"
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_host"

	"github.com/ccutch/congo/apps/launchpad/models"
)

type HostsController struct {
	congo.BaseController
	host *congo_host.CongoHost
}

func (hosts *HostsController) Setup(app *congo.Application) {
	hosts.Application = app
	app.HandleFunc("POST /hosts", hosts.handleCreate)

}

func (hosts HostsController) Handle(r *http.Request) congo.Controller {
	hosts.Request = r
	return &hosts
}

func (hosts *HostsController) CurrentHost() (*models.Host, error) {
	return models.GetHost(hosts.Application.DB, hosts.PathValue("host"))
}

func (hosts *HostsController) Searchhosts() ([]*models.Host, error) {
	return models.SearchHosts(hosts.DB, hosts.PathValue("query"))
}

func (hosts HostsController) handleCreate(w http.ResponseWriter, r *http.Request) {
	auth := hosts.Use("auth").(*congo_auth.Controller)
	i, _ := auth.Authenticate("user", r)
	if i == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	name, size, reg := r.FormValue("name"), r.FormValue("size"), r.FormValue("region")
	host, err := models.NewHost(hosts.DB, i.ID, name, size, reg)
	if err != nil {
		hosts.Render(w, r, "error-message", err)
		return
	}

	go func(host *models.Host) {
		storage := map[string]int64{"SM": 5, "MD": 25, "LG": 50, "XL": 100}[host.Size]
		server, err := hosts.host.NewServer(host.Name, host.Region, host.Size, storage)
		if err != nil {
			host.Error = err.Error()
		} else {
			host.IpAddr = server.IP
		}

		if err := host.Save(); err != nil {
			log.Println("Failed to save server", server, err)
		}
	}(host)

	hosts.Redirect(w, r, "/hosts/"+host.ID)
}
