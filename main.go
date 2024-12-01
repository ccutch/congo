package main

import (
	"cmp"
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/ccutch/congo/controllers"
	"github.com/ccutch/congo/pkg/congo"
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

	if cert, key := sllCerts(); cert != "" && key != "" {
		log.Print("Serving Congo Server @ https://localhost:443")
		if err := http.ListenAndServeTLS("0.0.0.0:443", cert, key, nil); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Print("Serving Congo Server @ http://localhost:" + port)
		if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
			log.Fatal(err)
		}
	}
}

func sllCerts() (string, string) {
	cert, key := "/root/fullchain.pem", "/root/privkey.pem"
	if _, err := os.Stat(cert); os.IsNotExist(err) {
		return "", ""
	}
	if _, err := os.Stat(key); os.IsNotExist(err) {
		return "", ""
	}
	return cert, key
}
