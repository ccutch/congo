package main

import (
	"cmp"
	"embed"
	"log"
	"net/http"
	"os"

	"congo.gitpost.app/controllers"
	"congo.gitpost.app/internal/congo"
)

var (
	//go:embed all:migrations
	migrations embed.FS

	//go:embed all:templates
	templates embed.FS

	port = cmp.Or(os.Getenv("PORT"), "5000")
	path = cmp.Or(os.Getenv("DATA_PATH"), os.TempDir())
	db   = congo.SetupDatabase(path, migrations)

	server = congo.NewServer(
		congo.WithDatabase(db),
		congo.WithController("posts", &controllers.PostController{}),
		congo.WithTemplates(templates),
	)
)

func main() {
	http.Handle("/", server.ServeMux)

	http.Handle("/{$}", server.Serve("homepage.html"))
	// http.Handle("/_/", server.GitServer())

	http.Handle("GET /blog", server.Serve("blog-posts.html"))
	http.Handle("GET /blog/write", server.Serve("write-post.html"))
	http.Handle("GET /blog/{post}", server.Serve("read-post.html"))
	http.Handle("GET /blog/{post}/edit", server.Serve("edit-post.html"))

	log.Println("Serving HTTP @ http://localhost:" + port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}
