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

type Application struct {
	DB          *Database
	controllers map[string]Controller
	templates   *template.Template
	endpoints   *http.ServeMux
	creds       *Credentials
}

type Credentials struct {
	fullchain string
	privkey   string
}

type ApplicationOpt func(*Application) error

func NewApplication(opts ...ApplicationOpt) *Application {
	app := Application{
		controllers: map[string]Controller{},
		endpoints:   http.NewServeMux(),
	}

	for _, opt := range opts {
		if err := opt(&app); err != nil {
			log.Fatal("Failed to setup Congo server:", err)
		}
	}

	return &app
}

func (app *Application) Serve(name string) (view View) {
	view.App = app
	if view.template = app.templates.Lookup(name); view.template == nil {
		log.Fatalf("Template %s not found", name)
	}
	return view
}

func (app Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.endpoints.ServeHTTP(w, r)
}

func (app *Application) Start() error {
	http.Handle("/", app.endpoints)
	go app.sslServer()
	addr := "0.0.0.0:" + cmp.Or(os.Getenv("PORT"), "5000")
	log.Print("Serving Unsecure Congo @ http://" + addr)
	return http.ListenAndServe(addr, nil)
}

func (app *Application) sslServer() {
	if app.creds == nil {
		return
	}
	cert, key := app.creds.fullchain, app.creds.privkey
	log.Print("Serving Secure Congo @ https://localhost:443")
	err := http.ListenAndServeTLS("0.0.0.0:443", cert, key, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func WithController(name string, ctrl Controller) ApplicationOpt {
	return func(app *Application) error {
		return app.WithController(name, ctrl)
	}
}

func (app *Application) WithController(name string, controller Controller) error {
	app.controllers[name] = controller
	controller.Setup(app)
	return nil
}

func WithDatabase(db *Database) ApplicationOpt {
	return func(app *Application) error {
		return app.WithDatabase(db)
	}
}

func (app *Application) WithDatabase(db *Database) error {
	app.DB = db
	return db.MigrateUp()
}

func WithTemplates(templates fs.FS) ApplicationOpt {
	return func(app *Application) error {
		return app.WithTemplates(templates)
	}
}

func (app *Application) WithTemplates(templates fs.FS, patterns ...string) error {
	funcs := template.FuncMap{
		"db":  func() *Database { return app.DB },
		"req": func() *http.Request { return nil },
		"host": func() string {
			if env := os.Getenv("HOME"); env != "/home/coder" {
				return ""
			}
			port := cmp.Or(os.Getenv("PORT"), "5000")
			return fmt.Sprintf("/workspace-cgk/proxy/%s", port)
		},
	}

	for name, ctrl := range app.controllers {
		funcs[name] = func() Controller { return ctrl }
	}

	app.templates = template.New("").Funcs(funcs)
	if tmpl, err := app.templates.ParseFS(templates, "templates/*.html"); err == nil {
		app.templates = tmpl
	}

	if tmpl, err := app.templates.ParseFS(templates, "templates/**/*.html"); err == nil {
		app.templates = tmpl
	}

	return nil
}

func WithEndpoint(path string, fn HandlerFunc) ApplicationOpt {
	return func(app *Application) error {
		app.HandleFunc(path, fn)
		return nil
	}
}

func (app *Application) HandleFunc(path string, fn HandlerFunc) {
	app.endpoints.Handle(path, Endpoint{
		App:  app,
		Func: fn,
	})
}

func WithCredentials(cert, key string) ApplicationOpt {
	return func(app *Application) error {
		app.creds = &Credentials{cert, key}
		return nil
	}
}
