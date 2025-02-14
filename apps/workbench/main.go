package main

import (
	"bytes"
	"cmp"
	"embed"
	"log"
	"net/http"
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

	home, _ = os.UserHomeDir()
	data    = cmp.Or(os.Getenv("DATA_PATH"), home+"/congo")

	auth = congo_auth.InitCongoAuth(data,
		congo_auth.WithCookieName("workbench-"),
		congo_auth.WithSetupView("setup.html", "/"),
		congo_auth.WithAccessView("developer", "login.html"))

	host = congo_host.InitCongoHost(data)
	code = congo_code.InitCongoCode(data)

	app = congo.NewApplication(templates,
		congo.WithHost(os.Getenv("CONGO_HOST_PREFIX")),
		congo.WithTheme(cmp.Or(os.Getenv("CONGO_THEME"), "dark")),
		congo.WithDatabase(congo.SetupDatabase(data, "workbench.db", migrations)),
		congo.WithController(auth.Controller()),
		congo.WithController(controllers.Settings()),
		congo.WithController(controllers.Hosting(host)),
		congo.WithController(controllers.Services(host)),
		congo.WithController(controllers.Coding(host, code)))
)

func init() {
	go func() {
		var buf bytes.Buffer

		host := host.Local()
		host.SetStdout(&buf)
		if err := host.Run("curl", "-sSL", "https://nixpacks.com/install.sh"); err != nil {
			log.Println("Error loading nix pack", err)
		}

		if err := host.Run("bash", "-c", buf.String()); err != nil {
			log.Println("Error loading nix pack", err)
		}
	}()
}

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
