package congo

import (
	"cmp"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
)

type Controller interface {
	OnMount(*Server) error
	OnRequest(r *http.Request) Controller
}

type BaseController struct {
	*Server
	*http.Request
}

func (app *BaseController) Mount(server *Server) error {
	app.Server = server
	return nil
}

func (app *BaseController) Atoi(s string, def int) int {
	str := app.Request.URL.Query().Get(s)
	str = cmp.Or(str, app.Request.FormValue(s))
	i, err := strconv.Atoi(str)
	if err != nil {
		return def
	}
	return i
}

func (*BaseController) Host() string {
	if env := os.Getenv("HOME"); env != "/home/coder" {
		return ""
	}
	port := cmp.Or(os.Getenv("PORT"), "5000")
	return fmt.Sprintf("/workspace-cgk/proxy/%s", port)
}

func (app *BaseController) Refresh(w http.ResponseWriter, r *http.Request) {
	if htmx := r.Header.Get("Hx-Request"); htmx == "true" {
		w.Header().Add("Hx-Refresh", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, r.URL.Path, http.StatusFound)
}

func (app *BaseController) Redirect(w http.ResponseWriter, r *http.Request, path string) {
	if htmx := r.Header.Get("Hx-Request"); htmx == "true" {
		w.Header().Add("Hx-Location", app.Host()+path)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, path, http.StatusFound)
}

func (app *BaseController) Render(s *Server, w http.ResponseWriter, r *http.Request, page string, data any) {
	funcs := template.FuncMap{
		"db": func() *Database { return s.Database },
		"host": func() string {
			if env := os.Getenv("HOME"); env != "/home/coder" {
				return ""
			}
			port := cmp.Or(os.Getenv("PORT"), "5000")
			return fmt.Sprintf("/workspace-cgk/proxy/%s", port)
		},
	}
	for name, ctrl := range s.controllers {
		funcs[name] = func() Controller { return ctrl.OnRequest(r) }
	}
	if err := app.Server.templates.Funcs(funcs).Execute(w, data); err != nil {
		app.Server.templates.ExecuteTemplate(w, "error-message", err)
	}
}
