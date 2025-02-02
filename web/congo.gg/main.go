package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/web/congo.gg/controllers"
)

var (
	home, _ = os.UserHomeDir()
	data    = cmp.Or(os.Getenv("DATA_PATH"), home+"/congo")

	//go:embed all:templates
	templates embed.FS

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "congo.db", nil)),
		congo.WithTemplates(templates),
		congo.WithHtmlTheme("synthwave"),
		congo.WithHostPrefix(os.Getenv("CONGO_HOST_PREFIX")),
		congo.WithController("hosting", new(controllers.HostingController)))
)

func main() {
	app.Handle("/", app.Serve("homepage.html"))

	app.StartFromEnv()
}
