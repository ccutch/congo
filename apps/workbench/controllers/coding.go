package controllers

import (
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_code"
)

type CodingController struct {
	congo.BaseController
	*congo_code.CongoCode
	Repo *congo_code.Repository
	Work *congo_code.Workspace
}

func (code *CodingController) Setup(app *congo.Application) {
	code.CongoCode = congo_code.InitCongoCode(app)

	code.BaseController.Setup(app)
}

func (code CodingController) Handle(req *http.Request) congo.Controller {
	code.Request = req
	return &code
}
