package congo

import (
	"context"
	"html/template"
	"log"
	"net/http"
)

type View struct {
	App      *Application
	template *template.Template
	Request  *http.Request
	Error    error
}

func (view View) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	view.Request = req.WithContext(context.TODO())
	funcs := template.FuncMap{
		"db":  func() *Database { return view.App.DB },
		"req": func() *http.Request { return req },
	}

	for name, ctrl := range view.App.controllers {
		funcs[name] = func() Controller { return ctrl.Handle(req) }
	}

	if view.Error = view.template.Funcs(funcs).Execute(w, view); view.Error != nil {
		log.Print("Error rendering view: ", view.Error)
		view.template.ExecuteTemplate(w, "error-message", view.Error)
	}
}
