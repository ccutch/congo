package congo_call

import (
	"encoding/json"
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

func (c *CongoCall) Controller(roles ...string) (string, congo.Controller) {
	return "congo_call", &Controller{CongoCall: c, roles: roles}
}

type Controller struct {
	*CongoCall
	congo.BaseController
	roles []string
}

func (c *Controller) Setup(app *congo.Application) {
	c.BaseController.Setup(app)
	if auth, ok := app.Use("auth").(*congo_auth.AuthController); ok {
		http.Handle("POST /_call", auth.ProtectFunc(c.createRoom, c.roles...))
		http.Handle("GET /_call/{room}", auth.ProtectFunc(c.getRoom, c.roles...))
		http.Handle("GET /_call/{room}/events", auth.ProtectFunc(c.joinRoom, c.roles...))
		http.Handle("PUT /_call/{room}", auth.ProtectFunc(c.updateRoom, c.roles...))
		http.Handle("DELETE /_call/{room}", auth.ProtectFunc(c.deleteRoom, c.roles...))
	}
}

func (c Controller) Handle(req *http.Request) congo.Controller {
	c.Request = req
	return &c
}

func (c *Controller) createRoom(w http.ResponseWriter, r *http.Request) {
	room, err := c.NewRoom(r.FormValue("name"))
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	json.NewEncoder(w).Encode(&room)
}

func (c *Controller) getRoom(w http.ResponseWriter, r *http.Request) {
	room, err := c.GetRoom(r.PathValue("room"))
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}
	json.NewEncoder(w).Encode(&room)
}

func (c *Controller) joinRoom(w http.ResponseWriter, r *http.Request) {
	room, err := c.GetRoom(r.PathValue("room"))
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}

	flusher, err := c.EventStream(w, r)
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}

	i, _ := c.Use("auth").(*congo_auth.AuthController).Authenticate(r, c.roles...)
	peer := room.Register(i, flusher)

	for {
		select {
		case <-r.Context().Done():
			return
		case offer := <-peer.offers:
			flusher("call-offer", offer)
		case answer := <-peer.answers:
			flusher("call-answer", answer)
		case candidate := <-peer.candidates:
			flusher("candidate", candidate)
		}
	}
}

func (c *Controller) updateRoom(w http.ResponseWriter, r *http.Request) {
	room, err := c.GetRoom(r.PathValue("room"))
	if err != nil {
		c.Render(w, r, "error-message", err)
		return
	}

	if name := r.FormValue("name"); name != "" {
		room.Name = name
		room.Save()
		return
	}

	if call_type := r.FormValue("offer"); call_type != "" {
		room.Offer = &CallOffer{Type: call_type, Sdp: r.FormValue("sdp")}
		room.SendCallOffer(call_type, r.FormValue("sdp"))
		return
	}

	if call_type := r.FormValue("answer"); call_type != "" {
		room.Answer = &CallAnswer{Type: call_type, Sdp: r.FormValue("sdp")}
		room.SendCallAnswer(call_type, r.FormValue("sdp"))
		return
	}

	if candidate := r.FormValue("candidate"); candidate != "" {
		room.SendCandidate(candidate)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) deleteRoom(w http.ResponseWriter, r *http.Request) {

}
