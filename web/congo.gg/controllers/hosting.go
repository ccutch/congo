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
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/subscription"
)

type HostingController struct {
	congo.BaseController
	auth   *congo_auth.CongoAuth
	host   *congo_host.CongoHost
	sell   *congo_sell.CongoSell
	binary string
}

func Hosting(auth *congo_auth.CongoAuth, host *congo_host.CongoHost, sell *congo_sell.CongoSell) (string, *HostingController) {
	hosting := HostingController{auth: auth, host: host, sell: sell}
	go func() {
		var err error
		if hosting.binary, err = apps.Build("workbench"); err != nil {
			log.Fatal("Failed to build workbench: ", err)
		}
	}()
	return "hosting", &hosting
}

func (hosting *HostingController) Setup(app *congo.Application) {
	hosting.BaseController.Setup(app)
	http.HandleFunc("POST /checkout", hosting.goToCheckout)
	http.HandleFunc("GET /callback/{host}", hosting.callback)
	http.HandleFunc("POST /host/{host}/restart", hosting.restartHost)
	http.HandleFunc("DELETE /host/{host}", hosting.deleteHost)
	http.HandleFunc("POST /host/{host}/retry-deployment", hosting.retryDeployment)
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
	users, err := hosting.auth.Search(hosting.URL.Query().Get("query"))
	if err != nil || len(users) == 0 {
		return nil, err
	}
	for i := range size {
		results[i] = []*congo_auth.Identity{}
	}
	i := 0
	for _, users := range users {
		for _, user := range users {
			results[i%size] = append(results[i%size], user)
			i++
		}
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

	i, _ := hosting.auth.Authenticate(r, "user", "admin")
	server, err := models.NewServer(hosting.DB, i.ID, r.FormValue("name"), "s-1vcpu-2gb")
	if err != nil {
		log.Println("failed to create new server:", err)
		hosting.Render(w, r, "error-message", err)
		return
	}

	if _, err = hosting.host.NewServer(server.ID, server.Size, "sfo2"); err != nil {
		log.Println("failed to create new host:", err)
		hosting.Render(w, r, "error-message", err)
		return
	}

	server.CheckoutURL = fmt.Sprintf("https://congo.gg/callback/%s?checkout_id={CHECKOUT_SESSION_ID}", server.ID)
	server.CheckoutURL, err = products[0].Checkout(server.CheckoutURL)
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	server.Save()
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

	host, err := hosting.host.GetServer(server.ID)
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		if err = host.Launch(host.Region, host.Size, 5); err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}

		server.Status = models.Launched
		server.Save()

		server.IpAddr = host.Addr()
		server.Status = models.Prepared
		server.Save()

		server.Domain = fmt.Sprintf("%s.congo.gg", server.ID)
		domain := host.Domain(server.Domain)
		if err = host.Assign(domain); err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}

		domain.Save()
		if err = domain.Verify("admin@" + domain.DomainName); err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}

		server.Status = models.Assigned
		server.Save()

		if err = host.Deploy(hosting.binary); err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}

		server.Status = models.Ready

		server.Save()
	}()

	http.Redirect(w, r, "/host/"+host.ID, http.StatusFound)
}

func (hosting HostingController) deleteHost(w http.ResponseWriter, r *http.Request) {
	server, err := models.GetServer(hosting.DB, r.PathValue("host"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		if host, err := hosting.host.GetServer(server.ID); err == nil {
			if err = host.Reload(); err == nil {
				if server.Domain != "" {
					if err = host.Remove(host.Domain(server.Domain)); err != nil {
						server.Error = err.Error()
						server.Save()
						return
					}
				}
			}

			if err = host.Delete(false); err != nil {
				server.Error = err.Error()
				server.Save()
				return
			}
		}

		if server.CheckoutID != "" {
			checkout, err := session.Get(server.CheckoutID, nil)
			if err != nil {
				server.Error = err.Error()
				server.Save()
				return
			}

			_, err = subscription.Cancel(checkout.Subscription.ID, nil)
			if err != nil {
				server.Error = err.Error()
				server.Save()
				return
			}
		}

		if err = server.Delete(); err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}
	}()

	server.Status = models.Destroyed
	server.Save()

	hosting.Redirect(w, r, "/")
}

func (hosting HostingController) restartHost(w http.ResponseWriter, r *http.Request) {
	server, err := models.GetServer(hosting.DB, r.PathValue("host"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		host, err := hosting.host.GetServer(server.ID)
		if err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}

		if err = host.Reload(); err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}

		if err = host.Restart(); err != nil {
			server.Error = err.Error()
			server.Save()
			return
		}

		server.Status = models.Ready
		server.Save()
	}()

	hosting.Redirect(w, r, "/host/"+server.ID)
}

func (hosting HostingController) retryDeployment(w http.ResponseWriter, r *http.Request) {
	server, err := models.GetServer(hosting.DB, r.PathValue("host"))
	if err != nil {
		hosting.Render(w, r, "error-message", err)
		return
	}

	server.Error = ""
	server.Save()

	go func() {
		host, err := hosting.host.GetServer(server.ID)
		if err != nil {
			host, err = hosting.host.NewServer(server.ID, server.Size, "sfo2")
			if err != nil {
				server.Error = err.Error()
				server.Save()
				return
			}
		}

		switch server.Status {
		case models.Paid:
			if err = host.Launch(host.Region, host.Size, 5); err != nil {
				host.Reload()
			}
			fallthrough

		case models.Launched:
			server.IpAddr = host.Addr()
			server.Status = models.Prepared
			server.Save()
			fallthrough

		case models.Prepared:
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

			server.Status = models.Assigned
			server.Save()
			fallthrough

		case models.Assigned:
			server.Domain = fmt.Sprintf("%s.congo.gg", server.ID)
			domain := host.Domain(server.Domain)
			if err = host.Assign(domain); err != nil {
				server.Error = err.Error()
				server.Save()
				return
			}

			domain.Save()
			if err = domain.Verify("admin@" + domain.DomainName); err != nil {
				server.Error = err.Error()
				server.Save()
				return
			}

			host.Restart()
			server.Status = models.Ready
			server.Save()
		}
	}()

	hosting.Refresh(w, r)
}
