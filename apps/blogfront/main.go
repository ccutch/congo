package main

import (
	"cmp"
	"embed"
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
		congo_auth.WithSigninView("admin", "admin-login.html"),
		congo_auth.WithSigninView("writer", "writer-login.html"))

	app = congo.NewApplication(
		congo.WithDatabase(congo.SetupDatabase(path, "blog.db", migrations)),
		congo.WithTemplates(templates),
		congo.WithHtmlTheme(cmp.Or(os.Getenv("CONGO_THEME"), "dim")),
		congo.WithController(auth.Controller()),
		congo.WithController("posts", new(controllers.PostController)))
)

func main() {
	auth := app.Use("auth").(*congo_auth.Controller)

	app.Handle("GET /{$}", app.Serve("homepage.html"))
	app.Handle("GET /admin", auth.Protect(app.Serve("admin.html"), "admin"))
	app.Handle("GET /blog", app.Serve("blog-posts.html"))
	app.Handle("GET /{post}", app.Serve("read-post.html"))
	app.Handle("GET /blog/write", auth.Protect(app.Serve("write-post.html"), "writer", "admin"))
	app.Handle("GET /blog/{post}/edit", auth.Protect(app.Serve("edit-post.html"), "writer", "admin"))

	app.StartFromEnv()
}
