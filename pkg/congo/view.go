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
	*Server
	template *template.Template
	Request  *http.Request
	Error    error
}

func (view View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	view.Request = r.WithContext(context.TODO())
	funcs := template.FuncMap{
		"db": func() *Database { return view.Database },
		"host": func() string {
			if env := os.Getenv("HOME"); env != "/home/coder" {
				return ""
			}
			port := cmp.Or(os.Getenv("PORT"), "5000")
			return fmt.Sprintf("/workspace-cgk/proxy/%s", port)
		},
	}

	for name, ctrl := range view.controllers {
		funcs[name] = func() Controller { return ctrl.OnRequest(r) }
	}

	if view.Error = view.template.Funcs(funcs).Execute(w, view); view.Error != nil {
		log.Println("view error", view.Error)
		view.template.ExecuteTemplate(w, "error-message", view.Error)
	}
}
