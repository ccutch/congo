package congo

import "net/http"

type Endpoint struct {
	Secure  bool
	Handler http.HandlerFunc
}

func (api Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.Handler.ServeHTTP(w, r)
}
