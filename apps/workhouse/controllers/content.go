package controllers

import (
	"cmp"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ccutch/congo/apps/workhouse/models"
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_code"
	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_host/platforms/digitalocean"
	"github.com/pkg/errors"
)

type ContentController struct {
	congo.BaseController

	Code *congo_code.CongoCode
	Host *congo_host.CongoHost
	Repo *congo_code.Repository

	proxies map[string]http.Handler
}

func (c *ContentController) Setup(app *congo.Application) {
	c.BaseController.Setup(app)

	c.proxies = map[string]http.Handler{}
	c.Code = congo_code.InitCongoCode(app.DB.Root)
	c.Host = congo_host.InitCongoHost(app.DB.Root)
	c.Repo, _ = c.Code.NewRepo("source", congo_code.WithName("Code"))

	auth := app.Use("auth").(*AuthController)
	http.Handle("/source/", c.Repo.Serve(auth.AuthController, "developer"))
	http.Handle("/coder/", auth.ProtectFunc(c.handleWorkspace, "developer"))
	http.Handle("/download", auth.ProtectFunc(c.downloadSource, "developer"))
	http.Handle("POST /_content/post", auth.ProtectFunc(c.publishPost, "developer"))
	http.Handle("POST /_content/launch", auth.ProtectFunc(c.launchServer, "developer"))
	http.Handle("DELETE /_content/host/{host}", auth.ProtectFunc(c.deleteHost, "developer"))

	settings := app.Use("settings").(*SettingsController)
	if key := settings.get("HOST_API_KEY"); key != "" {
		c.Host.WithAPI(digitalocean.NewClient(key))
	}
}

func (c ContentController) Handle(req *http.Request) congo.Controller {
	c.Request = req
	return &c
}

func (c *ContentController) Files() []*congo_code.Blob {
	branch := cmp.Or(c.URL.Query().Get("branch"), "master")
	blobs, _ := c.Repo.Blobs(branch, c.PathValue("path"))
	return blobs
}

func (c *ContentController) CurrentBranch() string {
	return cmp.Or(c.URL.Query().Get("branch"), "master")
}

func (c *ContentController) CurrentFile() (blob *congo_code.Blob, err error) {
	branch, path := c.CurrentBranch(), c.PathValue("path")
	if blob, err = c.Repo.Open(branch, path); err != nil {
		path = filepath.Join(path, "README.md")
		if blob, err = c.Repo.Open(branch, path); err != nil {
			return nil, nil
		}
	}
	return blob, err
}

func (c *ContentController) Posts() ([]*models.Post, error) {
	return models.AllPosts(c.DB)
}

func (c *ContentController) Hosts(ownerID string) ([]*models.Host, error) {
	i, _ := c.Use("auth").(*AuthController).Authenticate(c.Request, "developer")
	return models.HostsForOwner(c.DB, i.ID)
}

func (c ContentController) downloadSource(w http.ResponseWriter, r *http.Request) {
	branch := cmp.Or(r.URL.Query().Get("branch"), "master")
	path := cmp.Or(r.URL.Query().Get("path"), ".")
	binary, err := c.Repo.Build(branch, path)
	if err != nil {
		log.Println("Failed to build binary: ", err)
	}
	w.Header().Set("Content-Disposition", "attachment; filename=congo")
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, binary)
}

func (c ContentController) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	i, _ := c.Use("auth").(*AuthController).Authenticate(r, "developer")
	if h, ok := c.proxies[i.ID]; ok {
		h.ServeHTTP(w, r)
		return
	}
	if workspace, err := c.Code.GetWorkspace("workspace-" + i.ID); err == nil {
		c.proxies[i.ID] = workspace.Proxy("/coder/")
		c.proxies[i.ID].ServeHTTP(w, r)
	} else {
		c.Render(w, r, "not-found.html", nil)
	}
}

func (c ContentController) publishPost(w http.ResponseWriter, r *http.Request) {
	_, err := models.NewPost(c.DB, r.FormValue("title"), r.FormValue("content"))
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	c.Refresh(w, r)
}

func (c ContentController) launchServer(w http.ResponseWriter, r *http.Request) {
	settings := c.Use("settings").(*SettingsController)
	i, _ := c.Use("auth").(*AuthController).Authenticate(r, "developer")
	name := r.FormValue("name")
	h, err := models.NewHost(c.DB, i.ID, "", name, "")
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	go func() {
		h.ServerID = fmt.Sprintf("%s-%s", settings.Name(), h.ID)
		h.ServerID = strings.ReplaceAll(strings.ToLower(h.ServerID), " ", "-")
		defer h.Save()

		size, region := settings.HostSize(), settings.HostRegion()
		server, err := c.Host.NewServer(h.ServerID, size, region)
		if err != nil {
			h.Status = "failed"
			h.Error = err.Error()
			return
		}

		storage, _ := strconv.Atoi(settings.StorageSize())
		if err = server.Launch(region, size, int64(storage)); err != nil {
			h.Status = "failed"
			h.Error = errors.Wrap(err, "failed to launch server").Error()
			return
		}

		out, err := c.Repo.Build("master", ".")
		if err != nil {
			h.Status = "failed"
			h.Error = errors.Wrap(err, "failed to build binary").Error()
			return
		}

		if err = server.Deploy(out); err != nil {
			h.Status = "failed"
			h.Error = errors.Wrap(err, "failed to deploy binary").Error()
			return
		}

		if ns := settings.get("DOMAIN_ROOT"); ns != "" {
			h.DomainName = strings.ReplaceAll(strings.ToLower(name), " ", "-")
			h.DomainName = fmt.Sprintf("https://%s.%s", h.DomainName, ns)
		} else {
			h.DomainName = fmt.Sprintf("http://%s:8080", server.Addr())
		}

		h.Status = "ready"
	}()
	c.Refresh(w, r)
}

func (c ContentController) deleteHost(w http.ResponseWriter, r *http.Request) {
	h, err := models.GetHost(c.DB, r.PathValue("host"))
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}

	server, err := c.Host.GetServer(h.ServerID)
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}

	if err = server.Reload(); err != nil {
		c.Render(w, r, "error-message", err)
		return
	}

	if err = server.Delete(false); err != nil {
		c.Render(w, r, "error-message", err)
		return
	}

	if err = h.Delete(); err != nil {
		c.Render(w, r, "error-message", err)
		return
	}

	c.Refresh(w, r)
}
