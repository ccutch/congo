package main

import (
	"cmp"
	"embed"
	"net/http"
	"os"

	"github.com/ccutch/congo/apps/launchpad/controllers"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_boot"
	"github.com/ccutch/congo/pkg/congo_stat"
)

var (
	//go:embed all:migrations
	migrations embed.FS

	//go:embed all:templates
	templates embed.FS

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/launchpad")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithLoginView("login.html"),
		congo_auth.WithSetupView("setup.html"),
		congo_auth.WithDefaultRoles("admin", "user"),
		congo_auth.WithDefaultRole("user"))

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "app.db", migrations)),
		congo.WithHostPrefix("/coder/proxy/8000"),
		congo.WithController("auth", auth.Controller()),
		congo.WithController("apps", new(controllers.AppsController)),
		congo.WithController("hosts", new(controllers.HostsController)),
		congo.WithHtmlTheme("wireframe"),
		congo.WithTemplates(templates))
)

func main() {
	auth := app.Use("auth").(*congo_auth.Controller)

	http.Handle("/{$}", auth.Track(app.Serve("homepage.html"), "user"))
	http.Handle("/pricing", auth.Track(app.Serve("pricing.html"), "user"))
	http.Handle("/register", auth.Track(app.Serve("register.html"), "user"))

	http.Handle("/hosts", auth.Protect(app.Serve("my-hosts.html")))
	http.Handle("/hosts/{host}", auth.Protect(app.Serve("host-details.html")))
	http.Handle("/create/host", auth.Protect(app.Serve("create-host.html")))

	http.Handle("/apps", auth.Protect(app.Serve("apps-dashboard.html"), "admin"))
	http.Handle("/users", auth.Protect(app.Serve("users-dashboard.html"), "admin"))
	http.Handle("/create/app", auth.Protect(app.Serve("create-app.html"), "admin"))

	congo_boot.StartFromEnv(app, congo_stat.NewMonitor(app, auth))
}
