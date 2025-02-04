package main

import (
	"cmp"
	"embed"
	"net/http"
	"os"

	"github.com/ccutch/congo/pkg/congo"
)

var (
	//go:embed all:public
	public embed.FS

	//go:embed all:templates
	templates embed.FS

	app = congo.NewApplication(templates,
		congo.WithTheme(cmp.Or(os.Getenv("THEME"), "forest")),
		congo.WithHost(cmp.Or(os.Getenv("CONGO_HOST"), "")),
		congo.WithFunc("gallery", gallery))
)

func main() {
	app.Handle("GET /", app.Serve("homepage.html"))
	app.Handle("GET /gallery", app.Serve("gallery.html"))
	app.Handle("GET /public/", http.FileServerFS(public))

	app.StartFromEnv()
}
