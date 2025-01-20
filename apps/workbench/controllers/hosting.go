package controllers

import (
	"log"
	"net/http"
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

	auth := app.Use("auth").(*congo_auth.AuthController)
	hosting.host = congo_host.InitCongoHost(app.DB.Root, nil)

	app.HandleFunc("POST /_hosting/launch", auth.ProtectFunc(hosting.handleLaunch, "developer"))
	app.HandleFunc("POST /_hosting/restart/{server}", auth.ProtectFunc(hosting.handleRestart, "developer"))
	app.HandleFunc("POST /_hosting/domain", auth.ProtectFunc(hosting.handleDomain, "developer"))
	app.HandleFunc("POST /_hosting/verify/{domain}", auth.ProtectFunc(hosting.handleVerify, "developer"))
}

func (hosting HostingController) Handle(req *http.Request) congo.Controller {
	hosting.Request = req
	return &hosting
}

func (hosting *HostingController) List() ([]*congo_host.RemoteHost, error) {
	return hosting.host.ListServers()
}

func (hosting HostingController) handleLaunch(w http.ResponseWriter, r *http.Request) {
	storage, err := strconv.Atoi(r.FormValue("storage"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	name, region, size := r.FormValue("name"), r.FormValue("region"), r.FormValue("size")
	server, err := hosting.host.NewServer(name, size, region)
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	if err := server.Launch(region, size, int64(storage)); err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		if err := server.Prepare(); err != nil {
			// TODO: record error
			return
		}

		coding := hosting.Use("coding").(*CodingController)
		if source, err := coding.Repo.Build("master", "."); err != nil {
			log.Println("Failed to build binary: ", err)
		} else if err = server.Deploy(source); err != nil {
			log.Println("Failed to deploy server: ", err)
		}
	}()

	hosting.Refresh(w, r)
}

func (hosting HostingController) handleDomain(w http.ResponseWriter, r *http.Request) {
	server, err := hosting.host.GetServer(r.FormValue("server"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	if err := server.Reload(); err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	if domain := r.FormValue("domain"); domain != "" {
		d := server.Domain(domain)
		if err = d.Verify(); err == nil {
			hosting.Render(w, r, "error-message", err)
			return
		}
	}

	if err := server.Save(); err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	hosting.Refresh(w, r)
}

func (hosting HostingController) handleRestart(w http.ResponseWriter, r *http.Request) {
	server, err := hosting.host.GetServer(r.FormValue("server"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	if err := server.Reload(); err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		coding := hosting.Use("coding").(*CodingController)
		if source, err := coding.Repo.Build("master", "."); err != nil {
			log.Println("Failed to build binary: ", err)
		} else if err = server.Deploy(source); err != nil {
			log.Println("Failed to deploy server: ", err)
		}
	}()

	hosting.Refresh(w, r)
}

func (hosting HostingController) handleVerify(w http.ResponseWriter, r *http.Request) {
	domain, err := hosting.host.GetDomain(r.PathValue("domain"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	if err = domain.Verify(); err == nil {
		domain.Verified = true
		domain.Save()
	}

	hosting.Refresh(w, r)
}
