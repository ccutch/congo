package controllers

import (
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
)

type PaymentController struct {
	congo.BaseController
}

func (payments *PaymentController) Setup(app *congo.Application) {
	payments.BaseController.Setup(app)
}

func (payments PaymentController) Handle(req *http.Request) congo.Controller {
	payments.Request = req
	return &payments
}
