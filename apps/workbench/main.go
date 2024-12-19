package main

import (
	"cmp"
	"embed"
	"net/http"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_boot"
	"github.com/ccutch/congo/pkg/congo_code"
	"github.com/ccutch/congo/pkg/congo_stat"

	"github.com/ccutch/congo/apps/workbench/controllers"
)

var (
	//go:embed all:templates
	templates embed.FS

	//go:embed all:migrations
	migrations embed.FS

	path = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo-workbench")

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(path, "app.db", migrations)),
		congo.WithController("settings", new(controllers.SettingsController)),
		congo.WithTemplates(templates),
		congo.WithHtmlTheme("business"))

	auth = congo_auth.InitCongoAuth(app,
		congo_auth.WithDefaultRole("developer"),
		congo_auth.WithSetupView("setup.html"),
		congo_auth.WithLoginView("login.html"))

	code = congo_code.InitCongoCode(app,
		congo_code.WithGitServer(auth))
)

func main() {
	app.Handle("/", auth.Secure(app.Serve("workbench.html")))

	repo, _ := code.Repo("code", congo_code.WithName("Code"))
	app.Handle("/code/", repo)

	workspace, _ := code.Workspace("workspace", congo_code.WithRepo(repo))
	app.Handle("/coder/", auth.Secure(http.StripPrefix("/coder/", workspace)))

	congo_boot.StartFromEnv(app,
		congo_boot.IgnoreErr(workspace),
		congo_stat.NewMonitor(app, auth))
}
