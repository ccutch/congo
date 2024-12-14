package main

import (
	"cmp"
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/ccutch/congo/controllers"
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_boot"
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

	auth = congo_auth.OpenDirectory(app)
)

func main() {
	http.Handle("GET /{$}", app.Serve("homepage.html"))
	http.Handle("GET /admin", auth.Secure(app.Serve("admin.html")))

	http.Handle("GET /blog", app.Serve("blog-posts.html"))
	http.Handle("GET /blog/write", auth.Secure(app.Serve("write-post.html")))
	http.Handle("GET /blog/{post}", app.Serve("read-post.html"))
	http.Handle("GET /blog/{post}/edit", app.Serve("edit-post.html"))

	congo_boot.StartFromEnv(app, congo_stat.NewMonitor(app, auth))
}
