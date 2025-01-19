package main

import (
	"cmp"
	"embed"
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

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "launchpad.db", migrations)),
		congo.WithTemplates(templates),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_THEME"), "wireframe")),
		congo.WithController(auth.Controller()),
		congo.WithController("apps", new(controllers.AppsController)),
		congo.WithController("hosts", new(controllers.HostsController)))
)

func main() {
	auth := app.Use("auth").(*congo_auth.AuthController)

	app.Handle("/{$}", auth.Check(app.Serve("homepage.html"), "user"))
	app.Handle("/pricing", auth.Check(app.Serve("pricing.html"), "user"))
	app.Handle("/register", auth.Check(app.Serve("register.html"), "user"))

	app.Handle("/hosts", auth.Protect(app.Serve("hosts-dashboard.html")))
	app.Handle("/hosts/{host}", auth.Protect(app.Serve("host-details.html")))
	app.Handle("/create/host", auth.Protect(app.Serve("create-host.html")))

	app.Handle("/apps", auth.Protect(app.Serve("apps-dashboard.html"), "admin"))
	app.Handle("/users", auth.Protect(app.Serve("users-dashboard.html"), "admin"))
	app.Handle("/create/app", auth.Protect(app.Serve("create-app.html"), "admin"))

	app.StartFromEnv()
}
