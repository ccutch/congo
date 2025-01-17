package congo

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

func (ctrl *BaseController) EventStream(w http.ResponseWriter, r *http.Request) (func(string, any), error) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("event streaming not supported")
	}

	fmt.Fprintf(w, "event: ping\ndata: pong\n\n")
	flusher.Flush()

	return func(template string, data any) {
		var buf bytes.Buffer
		ctrl.Render(&buf, r, template, data)
		data = strings.ReplaceAll(buf.String(), "\n", "")
		if _, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", data); err != nil {
			log.Println("Failed to flush: ", template, data)
		}
		flusher.Flush()
	}, nil
}
