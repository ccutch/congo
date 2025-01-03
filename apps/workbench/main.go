package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_code"
	"github.com/ccutch/congo/pkg/congo_host"

	"github.com/ccutch/congo/apps/workbench/controllers"
)

var (
	//go:embed all:templates
	templates embed.FS

	//go:embed all:migrations
	migrations embed.FS

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo-workbench")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithDefaultRole("developer"),
		congo_auth.WithSetupView("setup.html"),
		congo_auth.WithSigninView("developer", "login.html"))

	host = congo_host.InitCongoHost(data)

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "app.db", migrations)),
		congo.WithController("auth", auth.Controller()),
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

	coding.Repo, _ = coding.Repository("code", congo_code.WithName("Code"))
	coding.Work = coding.Workspace("coder", 7000, coding.Repo)

	app.Handle("/", auth.Protect(app.Serve("workbench.html")))
	app.Handle("/code/", coding.Repo.Serve(auth, "developer"))
	app.Handle("/coder/", auth.Protect(coding.Work.Proxy("/coder/")))

	app.StartFromEnv(congo.Ignore("workspace", coding.Work))
}
