package congo

import "net/http"

type Endpoint struct {
	App  *Application
	Func HandlerFunc
}

type HandlerFunc func(*Application, http.ResponseWriter, *http.Request)

func (api Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.Func(api.App, w, r)
}
