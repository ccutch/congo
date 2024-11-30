package congo

import (
	"context"
	"html/template"
	"log"
	"net/http"
)

type View struct {
	*Server
	*template.Template
	Request *http.Request
	Error   error
}

func (view View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if view.Template == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	view.Request = r.WithContext(context.TODO())
	if view.Error = view.Template.Execute(w, view); view.Error != nil {
		log.Println("error", view.Error)
		view.templates.ExecuteTemplate(w, "error-message", view.Error)
	}
}
