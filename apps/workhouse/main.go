package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"

	"github.com/ccutch/congo/apps/workhouse/controllers"
)

var (
	//go:embed all:templates
	templates embed.FS

	//go:embed all:migrations
	migrations embed.FS

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithCookieName("workhouse"),
		congo_auth.WithSetupView("setup.html", "/"),
		congo_auth.WithAccessView("user", "welcome.html"),
		congo_auth.WithAccessView("developer", "signin.html"))

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "workhouse.db", migrations)),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("DAISY_THEME"), "dark")),
		congo.WithHostPrefix(os.Getenv("CONGO_HOST_PREFIX")),
		congo.WithTemplates(templates),
		congo.WithController("auth", controllers.NewAuthController(auth)),
		congo.WithController("settings", new(controllers.SettingsController)),
		congo.WithController("content", new(controllers.ContentController)))
)

func main() {
	auth := app.Use("auth").(*controllers.AuthController)

	app.Handle("/", app.Serve("not-found.html"))
	app.Handle("/signin", app.Serve("signin.html"))
	app.Handle("/{$}", auth.Serve("my-hosts.html", "user", "developer"))
	app.Handle("/code/{path...}", auth.Serve("our-code.html", "developer"))
	app.Handle("/users", auth.Serve("our-users.html", "developer"))
	app.Handle("/user/{user}", auth.Serve("our-users.html", "developer"))
	app.Handle("/settings", auth.Serve("settings.html", "developer"))
	app.Handle("/dev/{user}", auth.Serve("settings.html", "developer"))

	app.StartFromEnv()
}
