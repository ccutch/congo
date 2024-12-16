package congo

import (
	"cmp"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
)

type Application struct {
	DB          *Database
	controllers map[string]Controller
	endpoints   *http.ServeMux
	creds       *Credentials
	hostPrefix  string
	sources     []fs.FS
	templates   *template.Template
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
		sources:     []fs.FS{},
	}

	for _, opt := range opts {
		if err := opt(&app); err != nil {
			log.Fatal("Failed to setup Congo server:", err)
		}
	}

	return &app
}

func (app *Application) Template(name string) http.HandlerFunc {
	view := View{App: app}
	return func(w http.ResponseWriter, r *http.Request) {
		if view.template = app.templates.Lookup(name); view.template != nil {
			view.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	}
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

func (app *Application) WithTemplates(source fs.FS) error {
	app.sources = append(app.sources, source)
	return nil
}

func WithEndpoint(path string, fn http.HandlerFunc) ApplicationOpt {
	return func(app *Application) error {
		app.HandleFunc(path, fn)
		return nil
	}
}

func (app *Application) Handle(path string, fn http.Handler) {
	app.endpoints.Handle(path, Endpoint{
		App:  app,
		Func: fn.ServeHTTP,
	})
}

func (app *Application) HandleFunc(path string, fn http.HandlerFunc) {
	app.endpoints.Handle(path, Endpoint{
		App:  app,
		Func: fn,
	})
}

func (app *Application) WithCredentials(cert, key string) {
	if cert == "" || key == "" {
		return
	}
	app.creds = &Credentials{cert, key}
}

func WithHostPrefix(prefix string) ApplicationOpt {
	return func(app *Application) error {
		app.hostPrefix = prefix
		return nil
	}
}

func (app *Application) PrepareTemplates() {
	if app.templates != nil {
		return
	}

	funcs := template.FuncMap{
		"db":   func() *Database { return app.DB },
		"req":  func() *http.Request { return nil },
		"host": func() string { return app.hostPrefix },
	}

	for name, ctrl := range app.controllers {
		funcs[name] = func() Controller { return ctrl }
	}

	app.templates = template.New("").Funcs(funcs)

	for _, source := range app.sources {
		if tmpl, err := app.templates.ParseFS(source, "templates/*.html"); err == nil {
			app.templates = tmpl
		} else {
			log.Fatal("Failed to parse root templates", err)
		}

		if tmpl, err := app.templates.ParseFS(source, "templates/**/*.html"); err == nil {
			app.templates = tmpl
		} else {
			log.Print("Failed to parse root templates", err)
		}
	}
}
