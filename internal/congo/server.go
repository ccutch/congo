package congo

import (
	"cmp"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
)

type Server struct {
	controllers map[string]Controller
	database    *Database
	templates   *template.Template
	serveMux    *http.ServeMux
}

func NewServer(opts ...ServerOpt) *Server {
	server := Server{
		controllers: map[string]Controller{},
		serveMux:    http.NewServeMux(),
	}
	for _, opt := range opts {
		Must(opt(&server))
	}
	return &server
}

func (server *Server) Serve(name string) View {
	return View{
		Server:   server,
		Template: server.templates.Lookup(name),
	}
}

func (server Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.serveMux.ServeHTTP(w, r)
}

type ServerOpt func(*Server) error

func WithController(name string, ctrl Controller) ServerOpt {
	return func(server *Server) error {
		return server.WithController(name, ctrl)
	}
}

func (server *Server) WithController(name string, controller Controller) error {
	server.controllers[name] = controller
	return controller.Mount(server)
}

func WithDatabase(db *Database) ServerOpt {
	return func(server *Server) error {
		return server.WithDatabase(db)
	}
}

func (server *Server) WithDatabase(db *Database) error {
	log.Println("Loading data from", db.Root)
	server.database = db
	return db.MigrateUp()
}

func WithTemplates(templates fs.FS) ServerOpt {
	return func(server *Server) error {
		return server.WithTemplates(templates)
	}
}

func (server *Server) WithTemplates(templates fs.FS, patterns ...string) error {
	funcs := template.FuncMap{
		"db": func() *Database { return server.database },

		"host": func() string {
			if env := os.Getenv("HOME"); env != "/home/coder" {
				return ""
			}
			port := cmp.Or(os.Getenv("PORT"), "5000")
			return fmt.Sprintf("/workspace-cgk/proxy/%s", port)
		},
	}
	for name, ctrl := range server.controllers {
		funcs[name] = func() Controller { return ctrl }
	}
	server.templates = template.New("").Funcs(funcs)
	if tmpl, err := server.templates.ParseFS(templates, "templates/*.html"); err == nil {
		server.templates = tmpl
	}
	if tmpl, err := server.templates.ParseFS(templates, "templates/**/*.html"); err == nil {
		server.templates = tmpl
	}

	return nil
}

func WithEndpoint(path string, api Endpoint) ServerOpt {
	return func(server *Server) error {
		return server.WithEndpoint(path, api)
	}
}

func (server *Server) WithEndpoint(path string, api Endpoint) error {
	server.serveMux.Handle(path, api)
	return nil
}
