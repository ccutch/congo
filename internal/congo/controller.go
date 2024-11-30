package congo

import (
	"cmp"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type BaseController struct {
	*http.Request
	Error error
}

type Controller interface {
	Mount(*Server) error
	WithRequest(r *http.Request) Controller
}

func (app *BaseController) Atoi(s string, def int) (i int) {
	str := app.Request.URL.Query().Get(s)
	str = cmp.Or(str, app.Request.FormValue(s))
	if i, app.Error = strconv.Atoi(str); app.Error != nil {
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
