package congo_stat

import (
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
)

type Controller struct {
	congo.BaseController
	monitor *Monitor
}

func (m *Monitor) Controller() congo.Controller {
	return &Controller{monitor: m}
}

func (ctrl *Controller) Setup(app *congo.Application) {
	ctrl.BaseController.Setup(app)
	app.HandleFunc("GET /_stat/history", ctrl.handleStatHistory)
}

func (ctrl Controller) Handle(r *http.Request) congo.Controller {
	ctrl.Request = r
	return &ctrl
}
func (ctrl *Controller) handleStatHistory(w http.ResponseWriter, r *http.Request) {

}
