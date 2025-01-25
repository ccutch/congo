package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/ccutch/congo/apps/congo.gg/controllers"
	"github.com/ccutch/congo/pkg/congo"
)

var (
	//go:embed all:templates
	templates embed.FS

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo")

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "congo.db", nil)),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_THEME"), "synthwave")),
		congo.WithHostPrefix(os.Getenv("CONGO_HOST_PREFIX")),
		congo.WithTemplates(templates),
		congo.WithController("hosting", new(controllers.HostingController)),
		congo.WithController("payments", new(controllers.PaymentController)))
)

func main() {
	app.Handle("/", app.Serve("homepage.html"))

	app.StartFromEnv()
}
