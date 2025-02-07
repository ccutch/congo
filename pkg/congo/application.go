package congo

import (
	"cmp"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed all:templates
var congoTemplates embed.FS

type Application struct {
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

func NewApplication(templates fs.FS, opts ...ApplicationOpt) *Application {
	app := Application{
		controllers: map[string]Controller{},
		sources:     []fs.FS{congoTemplates, templates},
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
		log.Println("Serving: ", name, r.URL.Path)
		app.Render(w, r, name, nil)
	})
}

// Start runs the application HTTP server and SSL server
func (app *Application) Start() error {
	go app.sslServer()
	addr := "0.0.0.0:" + cmp.Or(os.Getenv("PORT"), "5000")
	log.Print("Serving Unsecure Congo @ http://" + addr)
	return http.ListenAndServe(addr, nil)
}

// sslServer starts the HTTPS server
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

// Use returns the controller with the given name
func (app Application) Use(name string) Controller {
	return app.controllers[name]
}

// WithController adds a controller to the application
func WithController(name string, ctrl Controller) ApplicationOpt {
	return func(app *Application) error {
		return app.WithController(name, ctrl)
	}
}

// WithController adds a controller to the application
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
		http.HandleFunc(path, fn)
		return nil
	}
}

func (app *Application) WithCredentials(cert, key string) {
	if cert == "" || key == "" {
		return
	}
	app.creds = &Credentials{cert, key}
}

func WithHost(prefix string) ApplicationOpt {
	return func(app *Application) error {
		app.hostPrefix = prefix
		return nil
	}
}

func WithTheme(theme string) ApplicationOpt {
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

func WithFunc(name string, fn any) ApplicationOpt {
	return func(app *Application) error {
		app.templates.Funcs(template.FuncMap{name: fn})
		return nil
	}
}

func (app *Application) PrepareTemplates() {
	funcs := template.FuncMap{
		"db":   func() *Database { return app.DB },
		"req":  func() *http.Request { return nil },
		"host": func() string { return app.hostPrefix },
		"raw":  func(val string) template.HTML { return template.HTML(val) },
		"path": func(parts ...string) string { return fmt.Sprintf("/%s", strings.Join(parts, "/")) },
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
			log.Println("Failed to parse templates:", err)
		}
	}
}

func (app *Application) Render(w io.Writer, r *http.Request, page string, data any) {
	funcs := template.FuncMap{
		"db":   func() *Database { return app.DB },
		"req":  func() *http.Request { return r },
		"host": func() string { return app.hostPrefix },
	}

	for name, ctrl := range app.controllers {
		funcs[name] = func() Controller { return ctrl.Handle(r) }
	}

	template := app.templates.Lookup(page)
	if template == nil {
		if rw, ok := w.(http.ResponseWriter); ok {
			http.Error(rw, "template not found", http.StatusInternalServerError)
		} else {
			fmt.Fprintf(w, "template not found")
		}
		return
	}

	if err := template.Funcs(funcs).Execute(w, data); err != nil {
		log.Print("Error rendering: ", err)
		app.templates.ExecuteTemplate(w, "error-message", err)
	}
}
