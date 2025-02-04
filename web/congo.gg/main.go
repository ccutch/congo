package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_host/platforms/digitalocean"
	"github.com/ccutch/congo/pkg/congo_sell"
	"github.com/ccutch/congo/pkg/congo_sell/backends/stripe"
	"github.com/ccutch/congo/web/congo.gg/controllers"
)

var (
	home, _ = os.UserHomeDir()
	data    = cmp.Or(os.Getenv("DATA_PATH"), home+"/congo")

	//go:embed all:templates
	templates embed.FS

	app = congo.NewApplication(templates,
		congo.WithHost(cmp.Or(os.Getenv("CONGO_HOST_PREFIX"), "")),
		congo.WithTheme(cmp.Or(os.Getenv("CONGO_THEME"), "synthwave")),
		congo.WithDatabase(congo.SetupDatabase(data, "congo.db", nil)),
		congo.WithController("hosting", &controllers.HostingController{

			Host: congo_host.InitCongoHost(data,
				congo_host.WithAPI(digitalocean.NewClient(os.Getenv("DIGITAL_OCEAN_API_KEY")))),

			Sell: congo_sell.InitCongoSell(data,
				congo_sell.WithBackend(stripe.NewClient(os.Getenv("STRIPE_KEY"))),
				congo_sell.WithProduct("Congo Workbench", "A Git driven development environment by Congo", 10000)),
		}))
)

func main() {
	app.Handle("/", app.Serve("homepage.html"))

	app.StartFromEnv()
}
