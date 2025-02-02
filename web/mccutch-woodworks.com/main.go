package main

import (
	"embed"
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
)

var (
	//go:embed all:public
	publicFiles embed.FS

	//go:embed all:templates
	templates embed.FS

	app = congo.NewApplication(
		congo.WithHtmlTheme("forest"),
		congo.WithHostPrefix("/coder/proxy/8000"),
		congo.WithTemplates(templates))
)

func main() {

	app.Handle("GET /", app.Serve("homepage.html"))
	app.Handle("GET /public/", http.FileServerFS(publicFiles))

	app.StartFromEnv()
}
