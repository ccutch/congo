package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/ccutch/congo/apps/congo.gg/controllers"
	"github.com/ccutch/congo/pkg/congo"
)

var (
	home, _ = os.UserHomeDir()
	data    = cmp.Or(os.Getenv("DATA_PATH"), home+"/congo")

	//go:embed all:templates
	templates embed.FS

	app = congo.NewApplication(
		congo.WithTemplates(templates),
		congo.WithHostPrefix(os.Getenv("CONGO_HOST_PREFIX")),
		congo.WithDatabase(congo.SetupDatabase(data, "congo.db", nil)),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_THEME"), "synthwave")),
		congo.WithController("hosting", new(controllers.HostingController)),
		congo.WithController("payments", new(controllers.PaymentController)))
)

func main() {
	app.Handle("/", app.Serve("homepage.html"))

	app.StartFromEnv()
}
