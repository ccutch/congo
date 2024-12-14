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
	Setup(*Application)
	Handle(*http.Request) Controller
}

type BaseController struct {
	*Application
	*http.Request
}

func (ctrl *BaseController) Setup(app *Application) {
	ctrl.Application = app
}

func (ctrl *BaseController) Atoi(s string, def int) int {
	str := ctrl.Request.URL.Query().Get(s)
	str = cmp.Or(str, ctrl.Request.FormValue(s))
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

func (ctrl *BaseController) Refresh(w http.ResponseWriter, r *http.Request) {
	if htmx := r.Header.Get("Hx-Request"); htmx == "true" {
		w.Header().Add("Hx-Refresh", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, r.URL.Path, http.StatusFound)
}

func (ctrl *BaseController) Redirect(w http.ResponseWriter, r *http.Request, path string) {
	if htmx := r.Header.Get("Hx-Request"); htmx == "true" {
		w.Header().Add("Hx-Location", ctrl.Host()+path)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, path, http.StatusFound)
}

func (ctrl *BaseController) Render(w http.ResponseWriter, r *http.Request, page string, data any) {
	funcs := template.FuncMap{
		"db":  func() *Database { return ctrl.DB },
		"req": func() *http.Request { return r },
		"host": func() string {
			if env := os.Getenv("HOME"); env != "/home/coder" {
				return ""
			}
			port := cmp.Or(os.Getenv("PORT"), "5000")
			return fmt.Sprintf("/workspace-cgk/proxy/%s", port)
		},
	}
	for name, ctrl := range ctrl.controllers {
		funcs[name] = func() Controller { return ctrl.Handle(r) }
	}
	if err := ctrl.Application.templates.Funcs(funcs).Execute(w, data); err != nil {
		ctrl.Application.templates.ExecuteTemplate(w, "error-message", err)
	}
}
