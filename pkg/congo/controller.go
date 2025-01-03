package congo

import (
	"cmp"
	"net/http"
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

func (ctrl *BaseController) Host() string {
	return ctrl.hostPrefix
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
