package congo_host

import (
	"embed"
	"net/http"
	"strconv"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

//go:embed all:templates
var Templates embed.FS

type Controller struct {
	congo.BaseController
	roles []string
	*CongoHost
}

func (host *CongoHost) Controller(roles ...string) (string, *Controller) {
	return "hosting", &Controller{CongoHost: host, roles: roles}
}

func (host *Controller) Setup(app *congo.Application) {
	host.BaseController.Setup(app)
	app.WithTemplates(Templates)
	if auth, ok := app.Use("auth").(*congo_auth.Controller); ok {
		app.Handle("POST /_host/launch", auth.ProtectFunc(host.launch, host.roles...))
	}
}

func (host *Controller) launch(w http.ResponseWriter, r *http.Request) {
	storage, err := strconv.Atoi(r.FormValue("storage"))
	if err != nil {
		host.Render(w, r, "error-message", err)
		return
	}
	name, region, size := r.FormValue("name"), r.FormValue("region"), r.FormValue("size")
	if _, err := host.NewServer(name, region, size, int64(storage)); err != nil {
		host.Render(w, r, "error-message", err)
		return
	}
	host.Refresh(w, r)
}

func (host Controller) Handle(r *http.Request) congo.Controller {
	host.Request = r
	return &host
}

func (host *Controller) Servers() ([]*Server, error) {
	return host.ListServers()
}
