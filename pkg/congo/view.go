package congo

import (
	"cmp"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

type View struct {
	App      *Application
	template *template.Template
	Request  *http.Request
	Error    error
}

func (view View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	view.Request = r.WithContext(context.TODO())
	funcs := template.FuncMap{
		"db":  func() *Database { return view.App.DB },
		"req": func() *http.Request { return r },
		"host": func() string {
			if env := os.Getenv("HOME"); env != "/home/coder" {
				return ""
			}
			port := cmp.Or(os.Getenv("PORT"), "5000")
			return fmt.Sprintf("/workspace-cgk/proxy/%s", port)
		},
	}

	for name, ctrl := range view.App.controllers {
		funcs[name] = func() Controller { return ctrl.Handle(r) }
	}

	if view.Error = view.template.Funcs(funcs).Execute(w, view); view.Error != nil {
		log.Println("view error", view.Error)
		view.template.ExecuteTemplate(w, "error-message", view.Error)
	}
}
