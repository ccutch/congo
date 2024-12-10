package main

import (
	"cmp"
	"embed"
	"net/http"
	"os"

	"github.com/ccutch/congo/controllers"
	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_run"
	"github.com/ccutch/congo/pkg/congo_stat"
)

var (
	//go:embed all:migrations
	migrations embed.FS

	//go:embed all:templates
	templates embed.FS

	path = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo-data")

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(path, "app.db", migrations)),
		congo.WithController("posts", &controllers.PostController{}),
		congo.WithTemplates(templates))

	dir = congo_auth.OpenDirectory(app)
)

func main() {
	http.Handle("GET /{$}", app.Serve("homepage.html"))
	// http.Handle("/code/", gitpost.Repo(path, "congo"))

	http.Handle("GET /blog", app.Serve("blog-posts.html"))
	http.Handle("GET /blog/write", app.Serve("write-post.html"))
	http.Handle("GET /blog/{post}", app.Serve("read-post.html"))
	http.Handle("GET /blog/{post}/edit", app.Serve("edit-post.html"))

	http.Handle("GET /admin", dir.Secure(app.Serve("admin.html")))

	congo_run.StartFromEnv(app, congo_stat.NewMonitor(app, dir))
}
