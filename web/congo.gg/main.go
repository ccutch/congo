package main

import (
	"cmp"
	"embed"
	"net/http"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_host/platforms/digitalocean"
	"github.com/ccutch/congo/pkg/congo_sell"
	"github.com/ccutch/congo/pkg/congo_sell/backends/stripe"
	"github.com/ccutch/congo/web/congo.gg/controllers"
)

var (
	home, _ = os.UserHomeDir()
	data    = cmp.Or(os.Getenv("DATA_PATH"), home+"/congo")

	//go:embed all:public
	public embed.FS

	//go:embed all:templates
	templates embed.FS

	//go:embed all:migrations
	migrations embed.FS

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithCookieName("congo-gg"),
		congo_auth.WithSetupView("setup.html", "/admin"),
		congo_auth.WithAccessView("user", "homepage.html"),
		congo_auth.WithAccessView("admin", "admin-login.html"),
		congo_auth.WithSigninDest("/"))

	host = congo_host.InitCongoHost(data,
		congo_host.WithPlatform(digitalocean.NewClient(os.Getenv("DIGITAL_OCEAN_API_KEY"))))

	sell = congo_sell.InitCongoSell(data,
		congo_sell.WithBackend(stripe.NewClient(os.Getenv("STRIPE_KEY"))),
		congo_sell.WithProduct("Congo Workbench", "A cloud hosted coding environment by Congo", 12_00))

	app = congo.NewApplication(templates,
		congo.WithDatabase(congo.SetupDatabase(data, "congo.db", migrations)),
		congo.WithTheme(cmp.Or(os.Getenv("CONGO_THEME"), "synthwave")),
		congo.WithHost(os.Getenv("CONGO_HOST")),
		congo.WithController(auth.Controller()),
		congo.WithController(controllers.Hosting(auth, host, sell)))
)

func main() {
	auth := app.Use("auth").(*congo_auth.AuthController)

	http.Handle("/", app.Serve("not-found.html"))
	http.Handle("/public/", http.FileServerFS(public))
	http.Handle("/login", app.Serve("user-login.html"))
	http.Handle("/admin", auth.Serve("admin.html", "admin"))
	http.Handle("/{$}", auth.Serve("dashboard.html", "user", "admin"))
	http.Handle("/host/{host}", auth.Serve("dashboard.html", "user", "admin"))

	app.StartFromEnv()
}
