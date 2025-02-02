package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/ccutch/congo/pkg/congo"
)

var (
	//go:embed all:public
	publicFiles embed.FS

	//go:embed all:templates
	templates embed.FS

	app = congo.NewApplication(
		congo.WithHostPrefix(os.Getenv("CONGO_HOST")),
		congo.WithHtmlTheme("forest"),
		congo.WithTemplates(templates),
		congo.WithFunc("gallery", listGalleryImages))
)

func main() {
	app.Handle("GET /", app.Serve("homepage.html"))
	app.Handle("GET /gallery", app.Serve("gallery.html"))
	app.Handle("GET /public/", http.FileServerFS(publicFiles))

	app.StartFromEnv()
}

func listGalleryImages() [4][]string {
	result := [4][]string{}
	gallery, err := fs.Sub(publicFiles, "public/gallery")
	if err != nil {
		log.Println("Gallery not found")
		return result
	}
	var files []string
	_ = fs.WalkDir(gallery, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, d.Name())
		return nil
	})
	length := len(files)
	if length == 0 {
		return result
	}
	partSize := (length + 3) / 4
	for i := 0; i < length; i += partSize {
		end := i + partSize
		if end > length {
			end = length
		}
		result[i] = files[i:end]
	}
	return result
}
