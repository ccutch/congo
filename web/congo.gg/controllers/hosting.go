package controllers

import (
	"fmt"
	"net/http"

	"github.com/ccutch/congo/apps"
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_sell"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type HostingController struct {
	congo.BaseController
	auth *congo_auth.CongoAuth
	host *congo_host.CongoHost
	sell *congo_sell.CongoSell
}

func Hosting(auth *congo_auth.CongoAuth, host *congo_host.CongoHost, sell *congo_sell.CongoSell) (string, *HostingController) {
	return "hosting", &HostingController{auth: auth, host: host, sell: sell}
}

func (hosting *HostingController) Setup(app *congo.Application) {
	hosting.BaseController.Setup(app)
	app.HandleFunc("POST /launch", hosting.launchServer)
	app.HandleFunc("GET /checkout", hosting.goToCheckout)
	app.HandleFunc("GET /{host}/callback", hosting.callback)
}

func (hosting HostingController) Handle(req *http.Request) congo.Controller {
	hosting.Request = req
	return &hosting
}

func (hosting *HostingController) MyHosts() ([]*congo_host.RemoteHost, error) {
	i, _ := hosting.auth.Authenticate(hosting.Request, "user", "admin")
	if i == nil {
		return nil, errors.New("identity not found")
	}
	return hosting.HostsFor(i.ID)
}

func (hosting *HostingController) HostsFor(id string) ([]*congo_host.RemoteHost, error) {
	return hosting.host.ListServers()
}

func (hosting *HostingController) CurrentHost() (*congo_host.RemoteHost, error) {
	if hosting.PathValue("host") == "" {
		return nil, nil
	}
	return hosting.host.GetServer(hosting.PathValue("host"))
}

func (hosting *HostingController) UserGrid(size int) ([][]*congo_auth.Identity, error) {
	results := make([][]*congo_auth.Identity, size)
	users, err := hosting.auth.SearchByRole("user", hosting.URL.Query().Get("query"))
	if err != nil || len(users) == 0 {
		return nil, err
	}
	for i, user := range users {
		results[i%size] = append(results[i%size], user)
	}
	return results, nil
}

func (hosting HostingController) goToCheckout(w http.ResponseWriter, r *http.Request) {
	products, err := hosting.sell.Products()
	if err != nil || len(products) == 0 {
		hosting.Render(w, r, "error-message", err)
		return
	}

	host, err := hosting.host.NewServer("congo-"+uuid.NewString(), "s-1vcpu-2gb", "sfo2")
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	url := fmt.Sprintf("https://congo.gg/%s/checkout", host.ID)
	url, err = products[0].Checkout(url)
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (hosting HostingController) callback(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("session_id")
	if id == "" {
		hosting.Render(w, r, "error-message", errors.New("no session id found"))
		return
	}
	// TODO: track payment
	host, err := hosting.host.GetServer(r.PathValue("host"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		host.Launch(host.Region, host.Size, 5)
		host.Prepare()
		out, _ := apps.Build("workbench")
		host.Deploy(out)
	}()

	http.Redirect(w, r, "/", http.StatusFound)
}

func (hosting HostingController) launchServer(w http.ResponseWriter, r *http.Request) {
	var (
		app    = r.FormValue("server-type")
		name   = fmt.Sprintf("demo-%s-%s", app, uuid.NewString())
		size   = "s-1vcpu-2gb"
		region = "sfo2"
	)

	server, err := hosting.host.NewServer(name, size, region)
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
