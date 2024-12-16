package congo

import "net/http"

type Endpoint struct {
	App  *Application
	Func http.HandlerFunc
}

func (api Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.Func(w, r)
}
