package main

import (
	"cmp"
	"embed"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_code"
	"github.com/ccutch/congo/pkg/congo_host"

	"github.com/ccutch/congo/apps/workhouse/controllers"
)

var (
	//go:embed all:templates
	templates embed.FS

	//go:embed all:migrations
	migrations embed.FS

	data = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo")

	settings = &controllers.SettingsController{}

	coding = &controllers.CodingController{
		Code: congo_code.InitCongoCode(data),
		Host: congo_host.InitCongoHost(data, nil)}

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithSignupCallback(coding.HandleNewSignup),
		congo_auth.WithSetupView("setup.html", "/"),
		congo_auth.WithSigninView("developer", "login.html"))

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(data, "workhouse.db", migrations)),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_THEME"), "dark")),
		congo.WithTemplates(templates),
		congo.WithController(auth.Controller()),
		congo.WithController("coding", coding),
		congo.WithController("settings", settings))
)

func main() {
	auth := app.Use("auth").(*congo_auth.Controller)

	app.Handle("/", auth.Protect(app.Serve("homepage.html")))

	app.StartFromEnv()
}
