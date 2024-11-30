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

	posts = controllers.PostController{Database: db}

	server = congo.NewServer(
		congo.WithDatabase(db),
		congo.WithController("posts", &posts),
		congo.WithTemplates(templates),
	)
)

func main() {
	http.Handle("/", server.Serve("homepage.html"))
	// http.Handle("/_/", server.GitServer())

	http.Handle("/blog", server.Serve("blog-posts.html"))
	http.Handle("/blog/{post}", server.Serve("post.html"))
	http.HandleFunc("POST /blog", posts.CreatePost)

	log.Println("Serving HTTP @ http://localhost:" + port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}
