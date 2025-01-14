package controllers

import (
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
)

type AppsController struct{ congo.BaseController }

func (apps *AppsController) Setup(app *congo.Application) {
	apps.Application = app

}

func (apps AppsController) Handle(r *http.Request) congo.Controller {
	apps.Request = r
	return &apps
}
