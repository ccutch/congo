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
	*Database
	auth        Authenticator
	controllers map[string]Controller
	templates   *template.Template
	endpoints   *http.ServeMux
}

func NewServer(opts ...ServerOpt) *Server {
	server := Server{
		controllers: map[string]Controller{},
		endpoints:   http.NewServeMux(),
	}
	for _, opt := range opts {
		Must(opt(&server))
	}
	return &server
}

func (server *Server) Serve(name string) (view View) {
	view.Server = server
	if view.template = server.templates.Lookup(name); view.template == nil {
		log.Fatalf("Template %s not found", name)
	}
	return view
}

func (server Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.endpoints.ServeHTTP(w, r)
}

type ServerOpt func(*Server) error

func WithController(name string, ctrl Controller) ServerOpt {
	return func(server *Server) error {
		return server.WithController(name, ctrl)
	}
}

func (server *Server) WithController(name string, controller Controller) error {
	server.controllers[name] = controller
	return controller.OnMount(server)
}

func WithDatabase(db *Database) ServerOpt {
	return func(server *Server) error {
		return server.WithDatabase(db)
	}
}

func (server *Server) WithDatabase(db *Database) error {
	log.Println("Loading data from", db.Root)
	server.Database = db
	return db.MigrateUp()
}

func WithTemplates(templates fs.FS) ServerOpt {
	return func(server *Server) error {
		return server.WithTemplates(templates)
	}
}

func (server *Server) WithTemplates(templates fs.FS, patterns ...string) error {
	funcs := template.FuncMap{
		"db": func() *Database { return server.Database },
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

func WithEndpoint(path string, secure bool, fn http.HandlerFunc) ServerOpt {
	return func(server *Server) error {
		return server.WithEndpoint(path, secure, fn)
	}
}

func (server *Server) WithEndpoint(path string, secure bool, fn http.HandlerFunc) error {
	if secure {
		fn = server.auth.Secure(fn)
	}
	server.endpoints.Handle(path, Endpoint{
		Server:  server,
		Handler: fn,
	})
	return nil
}

func (server *Server) Start(addr string) {
	http.Handle("/", server.endpoints)

	go func() {
		if cert, key := server.certs(); cert != "" && key != "" {
			log.Print("Serving Congo Server @ https://localhost:443")
			if err := http.ListenAndServeTLS("0.0.0.0:443", cert, key, nil); err != nil {
				log.Fatal(err)
			}
		}
	}()

	log.Print("Serving Congo Server @ http://" + addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func (server *Server) certs() (string, string) {
	cert, key := "/root/fullchain.pem", "/root/privkey.pem"
	if _, err := os.Stat(cert); err != nil {
		return "", ""
	}
	if _, err := os.Stat(key); err != nil {
		return "", ""
	}
	return cert, key
}
