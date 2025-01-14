package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"

	"github.com/ccutch/congo/apps/workbench/controllers"
)

var (
	//go:embed all:templates
	templates embed.FS

	//go:embed all:migrations
	migrations embed.FS

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithSetupView("setup.html", "/"),
		congo_auth.WithSigninView("developer", "login.html"))

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "workbench.db", migrations)),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_THEME"), "dark")),
		congo.WithTemplates(templates),
		congo.WithController(auth.Controller()),
		congo.WithController("coding", new(controllers.CodingController)),
		congo.WithController("hosting", new(controllers.HostingController)),
		congo.WithController("settings", new(controllers.SettingsController)))
)

func main() {
	auth := app.Use("auth").(*congo_auth.Controller)
	coding := app.Use("coding").(*controllers.CodingController)

	app.Handle("/", auth.Protect(app.Serve("workbench.html")))
	app.Handle("/code/", coding.Repo.Serve(auth, "developer"))
	app.Handle("/coder/", auth.Protect(coding.Workspace.Proxy("/coder/")))

	app.StartFromEnv()
}
