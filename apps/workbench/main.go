package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_host"

	"github.com/ccutch/congo/apps/workbench/controllers"
)

var (
	//go:embed all:templates
	templates embed.FS

	//go:embed all:migrations
	migrations embed.FS

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithDefaultRole("developer"),
		congo_auth.WithSetupView("setup.html"),
		congo_auth.WithSigninView("developer", "login.html"))

	host = congo_host.InitCongoHost(data)

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "workbench.db", migrations)),
		congo.WithController(auth.Controller()),
		congo.WithController(host.Controller()),
		congo.WithController("coding", new(controllers.CodingController)),
		congo.WithController("servers", new(controllers.ServersController)),
		congo.WithController("settings", new(controllers.SettingsController)),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_THEME"), "dark")),
		congo.WithTemplates(templates))
)

func main() {
	auth := app.Use("auth").(*congo_auth.Controller)
	coding := app.Use("coding").(*controllers.CodingController)

	app.Handle("/", auth.Protect(app.Serve("workbench.html")))
	app.Handle("/code/", coding.Repo.Serve(auth, "developer"))
	app.Handle("/coder/", auth.Protect(coding.Workspace.Proxy("/coder/")))

	app.StartFromEnv()
}
