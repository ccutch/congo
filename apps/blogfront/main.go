package main

import (
	"cmp"
	"embed"
	"net/http"
	"os"

	"github.com/ccutch/congo/apps/blogfront/controllers"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
)

var (
	//go:embed all:migrations
	migrations embed.FS

	//go:embed all:templates
	templates embed.FS

	path = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir()+"/congo")

	auth = congo_auth.InitCongoAuth(path,
		congo_auth.WithSetupView("admin-setup.html", "/admin"),
		congo_auth.WithAccessView("admin", "admin-login.html"),
		congo_auth.WithAccessView("writer", "writer-login.html"))

	app = congo.NewApplication(templates,
		congo.WithTheme(cmp.Or(os.Getenv("CONGO_THEME"), "dim")),
		congo.WithDatabase(congo.SetupDatabase(path, "blog.db", migrations)),
		congo.WithController(auth.Controller()),
		congo.WithController("posts", new(controllers.PostController)))
)

func main() {
	auth := app.Use("auth").(*congo_auth.AuthController)

	http.Handle("GET /{$}", app.Serve("homepage.html"))
	http.Handle("GET /admin", auth.Protect(app.Serve("admin.html"), "admin"))
	http.Handle("GET /blog", app.Serve("blog-posts.html"))
	http.Handle("GET /{post}", app.Serve("read-post.html"))
	http.Handle("GET /blog/write", auth.Protect(app.Serve("write-post.html"), "writer", "admin"))
	http.Handle("GET /blog/{post}/edit", auth.Protect(app.Serve("edit-post.html"), "writer", "admin"))

	app.StartFromEnv()
}
