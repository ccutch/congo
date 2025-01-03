package congo

import (
	"cmp"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed all:templates
var templates embed.FS

type Application struct {
	router      *http.ServeMux
	DB          *Database
	controllers map[string]Controller
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
		router:      http.NewServeMux(),
		controllers: map[string]Controller{},
		sources:     []fs.FS{templates},
	}
	for _, opt := range opts {
		if err := opt(&app); err != nil {
			log.Fatal("Failed to setup Congo server:", err)
		}
	}
	return &app
}

func (app *Application) Serve(name string) http.Handler {
	app.PrepareTemplates()
	if page := app.templates.Lookup(name); page == nil {
		log.Fatalf("Template %s not found", name)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.Render(w, r, name, nil)
	})
}

func (app Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.router.ServeHTTP(w, r)
}

func (app *Application) Start() error {
	http.Handle("/", app.router)
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

func (app Application) Use(name string) Controller {
	return app.controllers[name]
}

func WithController(name string, ctrl Controller) ApplicationOpt {
	return func(app *Application) error {
		return app.WithController(name, ctrl)
	}
}

func (app *Application) WithController(name string, controller Controller) error {
	if _, ok := app.controllers[name]; !ok {
		app.controllers[name] = controller
		controller.Setup(app)
	} else {
		log.Println(name, "already registered controller")
	}
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
	app.router.Handle(path, fn)
}

func (app *Application) HandleFunc(path string, fn http.HandlerFunc) {
	app.router.HandleFunc(path, fn)
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

func WithHtmlTheme(theme string) ApplicationOpt {
	return func(app *Application) error {
		if app.templates == nil {
			app.templates = template.New("")
		}
		app.templates = app.templates.Funcs(template.FuncMap{
			"theme": func() string { return theme },
		})
		return nil
	}
}

func (app *Application) PrepareTemplates() {
	funcs := template.FuncMap{
		"db":   func() *Database { return app.DB },
		"req":  func() *http.Request { return nil },
		"host": func() string { return app.hostPrefix },
		"raw":  func(val string) template.HTML { return template.HTML(val) },
	}

	for name, ctrl := range app.controllers {
		funcs[name] = func() Controller { return ctrl }
	}

	if app.templates == nil {
		app.templates = template.New("")
	}
	app.templates = app.templates.Funcs(funcs)

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

func (app *Application) Render(w http.ResponseWriter, r *http.Request, page string, data any) {
	funcs := template.FuncMap{
		"db":   func() *Database { return app.DB },
		"req":  func() *http.Request { return r },
		"host": func() string { return app.hostPrefix },
	}

	for name, ctrl := range app.controllers {
		funcs[name] = func() Controller { return ctrl.Handle(r) }
	}

	template := app.templates.Lookup(page)
	if err := template.Funcs(funcs).Execute(w, data); err != nil {
		log.Print("Error rendering: ", err)
		app.templates.ExecuteTemplate(w, "error-message", err)
	}
}
