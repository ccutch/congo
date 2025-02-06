package main

import (
	"cmp"
	"embed"
	"log"
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
		congo_sell.WithProduct("Congo Workbench", "A Git driven development environment by Congo", 100_00))

	app = congo.NewApplication(templates,
		congo.WithDatabase(congo.SetupDatabase(data, "congo.db", nil)),
		congo.WithTheme(cmp.Or(os.Getenv("CONGO_THEME"), "synthwave")),
		congo.WithHost(os.Getenv("CONGO_HOST")),
		congo.WithController(auth.Controller()),
		congo.WithController(controllers.Hosting(auth, host, sell)))
)

func main() {
	if os.Getenv("STRIPE_KEY") == "" {
		log.Fatal("STRIPE_KEY is required")
	}

	auth := app.Use("auth").(*congo_auth.AuthController)

	app.Handle("/", app.Serve("not-found.html"))
	app.Handle("/public/", http.FileServerFS(public))
	app.Handle("/login", app.Serve("user-login.html"))
	app.Handle("/admin", auth.Serve("admin.html", "admin"))
	app.Handle("/{$}", auth.Serve("dashboard.html", "user", "admin"))
	app.Handle("/{host}", auth.Serve("dashboard.html", "user", "admin"))

	app.StartFromEnv()
}
