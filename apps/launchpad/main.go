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

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithCookieName("launchpad"),
		congo_auth.WithSigninDest("/hosts"),
		congo_auth.WithSetupView("setup.html", "/apps"),
		congo_auth.WithAccessView("user", "login-user.html"),
		congo_auth.WithAccessView("admin", "login-admin.html"))

	app = congo.NewApplication(templates,
		congo.WithDatabase(congo.SetupDatabase(data, "launchpad.db", migrations)),
		congo.WithTheme(cmp.Or(os.Getenv("CONGO_THEME"), "wireframe")),
		congo.WithController(auth.Controller()),
		congo.WithController("apps", new(controllers.AppsController)),
		congo.WithController("hosts", new(controllers.HostsController)))
)

func main() {
	auth := app.Use("auth").(*congo_auth.AuthController)

	http.Handle("/{$}", auth.Check(app.Serve("homepage.html"), "user"))
	http.Handle("/pricing", auth.Check(app.Serve("pricing.html"), "user"))
	http.Handle("/register", auth.Check(app.Serve("register.html"), "user"))

	http.Handle("/hosts", auth.Protect(app.Serve("hosts-dashboard.html")))
	http.Handle("/hosts/{host}", auth.Protect(app.Serve("host-details.html")))
	http.Handle("/create/host", auth.Protect(app.Serve("create-host.html")))

	http.Handle("/apps", auth.Protect(app.Serve("apps-dashboard.html"), "admin"))
	http.Handle("/users", auth.Protect(app.Serve("users-dashboard.html"), "admin"))
	http.Handle("/create/app", auth.Protect(app.Serve("create-app.html"), "admin"))

	app.StartFromEnv()
}
