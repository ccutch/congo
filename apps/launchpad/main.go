package main

import (
	"cmp"
	"embed"
	"net/http"
	"os"

	"github.com/ccutch/congo/apps/launchpad/controllers"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

var (
	//go:embed all:migrations
	migrations embed.FS

	//go:embed all:templates
	templates embed.FS

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/launchpad")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithCookieName("launchpad"),
		congo_auth.WithSetupView("setup.html"),
		congo_auth.WithSetupDest("/apps"),
		congo_auth.WithSigninView("user", "login-user.html"),
		congo_auth.WithSigninView("admin", "login-admin.html"),
		congo_auth.WithSigninDest("/hosts"))

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "launcher.db", migrations)),
		congo.WithHostPrefix("/coder/proxy/8000"),
		congo.WithController(auth.Controller()),
		congo.WithController("apps", new(controllers.AppsController)),
		congo.WithController("hosts", new(controllers.HostsController)),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_THEME"), "wireframe")),
		congo.WithTemplates(templates))
)

func main() {
	auth := app.Use("auth").(*congo_auth.Controller)

	http.Handle("/{$}", auth.Track(app.Serve("homepage.html"), "user"))
	http.Handle("/pricing", auth.Track(app.Serve("pricing.html"), "user"))
	http.Handle("/register", auth.Track(app.Serve("register.html"), "user"))

	http.Handle("/hosts", auth.Protect(app.Serve("hosts-dashboard.html")))
	http.Handle("/hosts/{host}", auth.Protect(app.Serve("host-details.html")))
	http.Handle("/create/host", auth.Protect(app.Serve("create-host.html")))

	http.Handle("/apps", auth.Protect(app.Serve("apps-dashboard.html"), "admin"))
	http.Handle("/users", auth.Protect(app.Serve("users-dashboard.html"), "admin"))
	http.Handle("/create/app", auth.Protect(app.Serve("create-app.html"), "admin"))

	app.StartFromEnv()
}
