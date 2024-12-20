package main

import (
	"cmp"
	"embed"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_boot"
	"github.com/ccutch/congo/pkg/congo_code"
	"github.com/ccutch/congo/pkg/congo_host"
	"github.com/ccutch/congo/pkg/congo_stat"

	"github.com/ccutch/congo/apps/workbench/controllers"
)

var (
	//go:embed all:templates
	templates embed.FS

	//go:embed all:migrations
	migrations embed.FS

	path = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo-workbench")
	host = congo_host.InitCongoHost(filepath.Join(path, "servers"))

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(path, "app.db", migrations)),
		congo.WithController("hosting", host.Controller()),
		congo.WithController("coding", new(controllers.CodingController)),
		congo.WithController("settings", new(controllers.SettingsController)),
		congo.WithTemplates(templates),
		congo.WithHtmlTheme("business"))

	auth = congo_auth.InitCongoAuth(app,
		congo_auth.WithDefaultRole("developer"),
		congo_auth.WithSetupView("setup.html"),
		congo_auth.WithLoginView("login.html"))
)

func main() {

	coding := app.Use("coding").(*controllers.CodingController)
	coding.Repo, _ = coding.Repository("code", congo_code.WithName("Code"))
	coding.Work, _ = coding.Workspace("coder", congo_code.WithRepo(coding.Repo))

	app.Handle("/", auth.Secure(app.Serve("workbench.html")))
	app.Handle("/code/", coding.Repo)
	app.Handle("/coder/", auth.Secure(http.StripPrefix("/coder/", coding.Work)))

	congo_boot.StartFromEnv(app,
		congo_boot.IgnoreErr(coding.Work),
		congo_stat.NewMonitor(app, auth))
}
