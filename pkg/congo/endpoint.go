package congo

import "net/http"

type Endpoint struct {
	Server  *Server
	Handler HandlerFunc
}

func (api Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.Handler(api.Server, w, r)
}
