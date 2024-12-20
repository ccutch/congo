package congo_host

import (
	"net/http"
	"strconv"

	"github.com/ccutch/congo/pkg/congo"
)

type Controller struct {
	congo.BaseController
	host *CongoHost
}

func (c *CongoHost) Controller() *Controller {
	return &Controller{host: c}
}

func (c *Controller) Setup(app *congo.Application) {
	c.BaseController.Setup(app)
	app.HandleFunc("POST /_host/launch", c.launchServer)
}

func (c Controller) Handle(req *http.Request) congo.Controller {
	c.Request = req
	return &c
}

func (c *Controller) Servers() ([]*Server, error) {
	return c.host.ListServers()
}

func (c Controller) launchServer(w http.ResponseWriter, r *http.Request) {
	storage, err := strconv.Atoi(r.FormValue("storage"))
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	name, region, size := r.FormValue("name"), r.FormValue("region"), r.FormValue("size")
	server, err := c.host.NewServer(name, region, size, int64(storage))
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	go server.Start()
	c.Refresh(w, r)
}
