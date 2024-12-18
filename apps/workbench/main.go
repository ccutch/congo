package main

import (
	"cmp"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ccutch/congo/example/controllers"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_boot"
	"github.com/ccutch/congo/pkg/congo_code"
	"github.com/ccutch/congo/pkg/congo_stat"
)

var (
	//go:embed all:migrations
	migrations embed.FS

	//go:embed all:templates
	templates embed.FS

	port = cmp.Or(os.Getenv("PORT"), "5000")
	path = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo-data")

	app = congo.NewApplication(
		congo.WithHostPrefix(fmt.Sprintf("/workspace-cgk/proxy/%s", port)),
		congo.WithDatabase(congo.SetupDatabase(path, "app.db", migrations)),
		congo.WithController("posts", new(controllers.PostController)),
		congo.WithTemplates(templates))

	auth = congo_auth.InitCongoAuth(app)

	code = congo_code.InitCongoCode(app,
		congo_code.WithGitServer(auth))

	repo, _ = code.Repo("code",
		congo_code.WithName("Code"))

	workspace, err = code.Workspace("workspace2",
		congo_code.WithRepo(repo),
		congo_code.WithPort(5001))
)

func main() {
	if err != nil {
		log.Println("Failed to setup workspace", err)
	}

	go func() {
		time.Sleep(time.Second)
		if err = workspace.Start(); err != nil {
			log.Println("Failed to start workspace", err)
		}
	}()

	app.Handle("GET /{$}", app.Serve("homepage.html"))
	app.Handle("/code/", repo.Server())

	coder := auth.Secure(workspace.Server())
	app.Handle("/coder/", http.StripPrefix("/coder/", coder))

	congo_boot.StartFromEnv(app, congo_stat.NewMonitor(app, auth))
}
