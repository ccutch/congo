package congo_host

import (
	"embed"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/google/uuid"
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

	if auth, ok := app.Use("auth").(*congo_auth.Controller); ok && len(host.roles) > 0 {
		app.Handle("POST /_host/launch", auth.ProtectFunc(host.launch, host.roles...))
		app.Handle("POST /_host/restart/{server}", auth.ProtectFunc(host.restart, host.roles...))
	}
}

func (host Controller) Handle(r *http.Request) congo.Controller {
	host.Request = r
	return &host
}

func (host *Controller) Servers() ([]*Server, error) {
	return host.ListServers()
}

func (host *Controller) launch(w http.ResponseWriter, r *http.Request) {
	storage, err := strconv.Atoi(r.FormValue("storage"))
	if err != nil {
		host.Render(w, r, "error-message", err)
		return
	}
	name, region, size := r.FormValue("name"), r.FormValue("region"), r.FormValue("size")
	if s, err := host.NewServer(name, region, size, int64(storage)); err != nil {
		host.Render(w, r, "error-message", err)
	} else {
		go s.Setup()
		host.Refresh(w, r)
	}
}

func (host *Controller) restart(w http.ResponseWriter, r *http.Request) {
	server, err := host.LoadServer(r.PathValue("server"))
	if err != nil {
		host.Render(w, r, "error-message", err)
		return
	}

	if src, _, err := r.FormFile("source"); err == nil {
		defer src.Close()

		root := filepath.Join(host.DB.Root, "host/files")
		os.MkdirAll(root, os.ModePerm)

		filePath := filepath.Join(root, uuid.NewString())
		dst, err := os.Create(filePath)
		if err != nil {
			host.Render(w, r, "error-message", err)
			return
		}

		defer dst.Close()
		if _, err = io.Copy(dst, src); err != nil {
			host.Render(w, r, "error-message", err)
			return
		}

		if err = server.Deploy(filePath); err != nil {
			host.Render(w, r, "error-message", err)
			return
		}
	}

	if server.Start(); server.Error != nil {
		host.Render(w, r, "error-message", server.Error)
		return
	}
	host.Refresh(w, r)
}
