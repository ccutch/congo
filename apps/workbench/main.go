package main

import (
	"cmp"
	"embed"
	"net/http"
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
		congo_auth.WithCookieName("workbench-"),
		congo_auth.WithSetupView("setup.html", "/"),
		congo_auth.WithAccessView("developer", "login.html"))

	app = congo.NewApplication(templates,
		congo.WithDatabase(congo.SetupDatabase(data, "workbench.db", migrations)),
		congo.WithTheme(cmp.Or(os.Getenv("CONGO_THEME"), "dark")),
		congo.WithHost(os.Getenv("CONGO_HOST_PREFIX")),
		congo.WithController(auth.Controller()),
		congo.WithController("coding", new(controllers.CodingController)),
		congo.WithController("hosting", new(controllers.HostingController)),
		congo.WithController("settings", new(controllers.SettingsController)))
)

func main() {
	auth := app.Use("auth").(*congo_auth.AuthController)
	coding := app.Use("coding").(*controllers.CodingController)

	http.Handle("/", auth.Serve("workbench.html", "developer"))
	http.Handle("/code/", coding.Repo.Serve(auth, "developer"))
	if coding.Workspace != nil {
		http.Handle("/coder/", auth.Protect(coding.Workspace.Proxy("/coder/"), "developer"))
	}

	app.StartFromEnv()
}
