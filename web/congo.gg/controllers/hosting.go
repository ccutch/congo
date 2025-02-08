package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ccutch/congo/apps"
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_sell"
	"github.com/ccutch/congo/web/congo.gg/models"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v81/checkout/session"
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
	http.HandleFunc("POST /launch", hosting.launchServer)
	http.HandleFunc("GET /checkout", hosting.goToCheckout)
	http.HandleFunc("GET /callback/{host}", hosting.callback)
	http.HandleFunc("DELETE /host/{host}", hosting.deleteHost)
}

func (hosting HostingController) Handle(req *http.Request) congo.Controller {
	hosting.Request = req
	return &hosting
}

func (hosting *HostingController) MyHosts() ([]*models.Server, error) {
	i, _ := hosting.auth.Authenticate(hosting.Request, "user", "admin")
	if i == nil {
		return nil, errors.New("identity not found")
	}
	return hosting.HostsFor(i.ID)
}

func (hosting *HostingController) HostsFor(id string) ([]*models.Server, error) {
	return models.ServersForUser(hosting.DB, id)
}

func (hosting *HostingController) CurrentHost() (*models.Server, error) {
	if hosting.PathValue("host") == "" {
		return nil, nil
	}
	return models.GetServer(hosting.DB, hosting.PathValue("host"))
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
		log.Println("failed to get products:", err)
		hosting.Render(w, r, "error-message", err)
		return
	}

	host, err := hosting.host.NewServer("congo-"+uuid.NewString(), "s-1vcpu-2gb", "sfo2")
	if err != nil {
		log.Println("failed to create new host:", err)
		hosting.Render(w, r, "error-message", err)
		return
	}

	i, _ := hosting.auth.Authenticate(r, "user", "admin")
	server, err := models.NewServer(hosting.DB, i.ID, host.ID, host.Name, host.Size)
	if err != nil {
		log.Println("failed to create new server:", err)
		hosting.Render(w, r, "error-message", err)
		return
	}

	server.CheckoutURL = fmt.Sprintf("https://congo.gg/callback/%s?checkout_id={CHECKOUT_SESSION_ID}", server.ID)
	server.CheckoutURL, err = products[0].Checkout(server.CheckoutURL)
	if err != nil {
		log.Println("failed to checkout:", err)
		hosting.Render(w, r, "error-message", err)
		return
	}

	server.Save()
	log.Println("checkout url:", server.CheckoutURL)
	http.Redirect(w, r, server.CheckoutURL, http.StatusFound)
}

func (hosting HostingController) callback(w http.ResponseWriter, r *http.Request) {
	server, err := models.GetServer(hosting.DB, r.PathValue("host"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	server.CheckoutID = r.URL.Query().Get("checkout_id")
	if server.CheckoutID == "" {
		hosting.Render(w, r, "error-message", errors.New("no session id found"))
		return
	}

	if _, err := session.Get(server.CheckoutID, nil); err != nil {
		hosting.Render(w, r, "error-message", fmt.Errorf("failed to get session: %s", err))
		return
	}

	server.Status = models.Paid
	server.Save()

	host, err := hosting.host.GetServer(server.HostID)
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		host.Launch(host.Region, host.Size, 5)

		server.Status = models.Launched
		server.Save()

		if err = host.Prepare(); err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}

		server.Status = models.Prepared
		server.Save()

		out, err := apps.Build("workbench")
		if err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}

		if err = host.Deploy(out); err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}

		server.IpAddr = host.Addr()
		server.Status = models.Ready
		server.Save()
	}()

	http.Redirect(w, r, "/host/"+host.ID, http.StatusFound)
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

func (hosting HostingController) deleteHost(w http.ResponseWriter, r *http.Request) {
	server, err := models.GetServer(hosting.DB, r.PathValue("host"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	host, err := hosting.host.GetServer(server.HostID)
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		if err = host.Reload(); err != nil {
			return
		}

		if err = host.Delete(true, false); err != nil {
			return
		}
	}()

	if err = server.Delete(); err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	hosting.Redirect(w, r, "/")
}
