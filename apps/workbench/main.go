package main

import (
	"cmp"
	"embed"
	"log"
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

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo-workbench")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithDefaultRole("developer"),
		congo_auth.WithSetupView("setup.html"),
		congo_auth.WithLoginView("login.html"))

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "app.db", migrations)),
		congo.WithController("auth", auth.Controller()),
		congo.WithController("coding", new(controllers.CodingController)),
		congo.WithController("hosting", new(controllers.HostingController)),
		congo.WithController("settings", new(controllers.SettingsController)),
		congo.WithHtmlTheme("dark"),
		congo.WithTemplates(templates))
)

func main() {
	auth := app.Use("auth").(*congo_auth.Controller)
	coding := app.Use("coding").(*controllers.CodingController)

	log.Println("coding", coding)
	coding.Repo = __(coding.Repository("code", congo_code.WithName("Code")))
	coding.Work = __(coding.Workspace("coder", congo_code.WithRepo(coding.Repo)))

	app.Handle("/", auth.Protect(app.Serve("workbench.html")))
	app.Handle("/code/", coding.Repo.Serve(auth, "developer"))
	app.Handle("/coder/", auth.Protect(http.StripPrefix("/coder/", coding.Work)))

	congo_boot.StartFromEnv(app,
		congo_boot.IgnoreErr(coding.Work),
		congo_stat.NewMonitor(app, auth))
}

func __[T any](val T, err error) T {
	if err != nil {
		log.Fatal("Ran into an error:", err)
	}
	log.Println(val)
	return val
}
