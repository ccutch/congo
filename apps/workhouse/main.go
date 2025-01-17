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

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "workhouse.db", migrations)),
		congo.WithController("auth", new(controllers.AuthController)),
		congo.WithController("content", new(controllers.ContentController)),
		congo.WithController("settings", new(controllers.SettingsController)),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_THEME"), "dark")),
		congo.WithTemplates(templates))
)

func main() {
	auth := app.Use("auth").(*congo_auth.Controller)

	app.Handle("/signin", app.Serve("signin.html"))

	app.Handle("/{$}", auth.Serve("homepage.html", "user", "developer"))
	app.Handle("/posts", auth.Serve("our-posts.html", "developer"))
	app.Handle("/users", auth.Serve("our-users.html", "developer"))
	app.Handle("/settings", auth.Serve("settings.html", "developer"))

	app.StartFromEnv()
}
